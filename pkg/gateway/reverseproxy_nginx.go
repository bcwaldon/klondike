package gateway

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var (
	nginxTemplateData = `
pid {{ .NGINXConfig.PIDFile }};
daemon on;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
{{ range $srv := $.ReverseProxyConfig.HTTPServers }}
    server {
        listen {{ $srv.ListenPort }};
        {{ if $srv.Name }}server_name {{ $srv.Name }}{{ if $srv.AltNames }} {{ range $srv.AltNames }}{{ . }}{{ end }}{{ end }};{{ end }}
        {{ if $srv.StaticCode -}}
        return {{ $srv.StaticCode }}{{ if $srv.StaticMessage }} '{{ $srv.StaticMessage }}'{{ end }};
        {{- else -}}
{{ range $loc := $srv.Locations }}
        location {{ or $loc.Path "/" }} {
            {{ if $loc.StaticCode -}}
			return {{ $loc.StaticCode }}{{ if $loc.StaticMessage }} '{{ $loc.StaticMessage }}'{{end}};
			{{- else }}
            proxy_pass http://{{ $loc.Upstream }};
			{{- end }}
        }
{{ end }}
{{ end }}
    }
{{ end }}
{{ range $up := $.ReverseProxyConfig.HTTPUpstreams }}
    upstream {{ $up.Name }} {
{{ range $ep := $up.Servers }}
        server {{ $ep.Host }}:{{ $ep.Port }};  # {{ $ep.Name }}
{{- end }}
    }
{{ end }}
}


stream {
{{ range $srv := $.ReverseProxyConfig.TCPServers }}
    server {
        listen {{ $srv.ListenPort }};
        proxy_pass {{ $srv.Upstream }};
    }
{{ end }}

{{ range $up := $.ReverseProxyConfig.TCPUpstreams }}
    upstream {{ $up.Name }} {
{{ range $ep := $up.Servers }}
        server {{ $ep.Host }}:{{ $ep.Port }};  # {{ $ep.Name }}
{{- end }}
    }

{{ end }}
}
`

	nginxTemplate = template.Must(template.New("nginx").Parse(nginxTemplateData))

	DefaultNGINXConfig = NGINXConfig{
		ClusterZone: "example.com",
		ConfigFile:  "/etc/nginx/nginx.conf",
		PIDFile:     "/var/run/nginx.pid",
		HealthPort:  7332,
		ListenPort:  7331,
	}
)

const (
	nginxStatusRunning = "running"
	nginxStatusStopped = "stopped"
	nginxStatusUnknown = "unknown"
)

type NGINXConfig struct {
	ClusterZone string
	ConfigFile  string
	PIDFile     string
	HealthPort  int
	ListenPort  int
}

func newNGINXConfig(hp int, cz string) NGINXConfig {
	cfg := DefaultNGINXConfig
	cfg.HealthPort = hp
	cfg.ClusterZone = cz
	return cfg
}

type NGINXManager interface {
	Status() (string, error)
	WriteConfig(*ServiceMap) error
	Start() error
	Reload() error
}

func newNGINXManager(cfg NGINXConfig) NGINXManager {
	return &nginxManager{cfg: cfg}
}

type nginxManager struct {
	cfg NGINXConfig
}

func (n *nginxManager) Status() (string, error) {
	log.Printf("Checking status")
	if _, err := os.Stat(n.cfg.PIDFile); err != nil {
		if os.IsNotExist(err) {
			return nginxStatusStopped, nil
		} else {
			return nginxStatusUnknown, err
		}
	}

	return nginxStatusRunning, nil
}

func (n *nginxManager) WriteConfig(sm *ServiceMap) error {
	cfg, err := renderConfig(&n.cfg, sm)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(n.cfg.ConfigFile, cfg, os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

func (n *nginxManager) assertConfigOK() error {
	return n.run("-t")
}

func (n *nginxManager) Start() error {
	log.Printf("Starting nginx")
	return n.run()
}

func (n *nginxManager) Reload() error {
	log.Printf("Reloading nginx")
	return n.run("-s", "reload")
}

func (n *nginxManager) run(args ...string) error {
	args = append([]string{"-c", n.cfg.ConfigFile}, args...)
	log.Printf("Calling run on nginx with args: %q", args)
	err := exec.Command("nginx", args...).Run()
	if err != nil {
		log.Printf("nginx command failed w/ err: %v", err)
		return err
	} else {
		log.Printf("nginx command success")
	}
	return nil
}

func newReverseProxyConfig(cfg *NGINXConfig, sm *ServiceMap) *reverseProxyConfig {
	data := reverseProxyConfig{
		HTTPServers: []httpReverseProxyServer{
			httpReverseProxyServer{
				ListenPort: cfg.HealthPort,
				Locations: []httpReverseProxyLocation{
					httpReverseProxyLocation{
						Path:          "/health",
						StaticCode:    200,
						StaticMessage: "Healthy!",
					},
				},
			},
			httpReverseProxyServer{
				ListenPort: cfg.ListenPort,
				StaticCode: 444,
			},
		},
	}

	for _, svg := range sm.HTTPServiceGroups {
		srv := httpReverseProxyServer{
			Name:       CanonicalHostname(svg, cfg.ClusterZone),
			AltNames:   svg.Aliases,
			ListenPort: cfg.ListenPort,
			Locations:  []httpReverseProxyLocation{},
		}

		for _, svc := range svg.Services {
			up := httpReverseProxyUpstream{
				Name:    strings.Join([]string{svc.Namespace, svg.Name(), svc.Name}, "__"),
				Servers: []httpReverseProxyUpstreamServer{},
			}

			for _, ep := range svc.Endpoints {
				up.Servers = append(up.Servers, httpReverseProxyUpstreamServer{
					Name: ep.Name,
					Host: ep.IP,
					Port: ep.Port,
				})
			}

			data.HTTPUpstreams = append(data.HTTPUpstreams, up)

			srv.Locations = append(srv.Locations, httpReverseProxyLocation{
				Path:     svc.Path,
				Upstream: up.Name,
			})
		}

		data.HTTPServers = append(data.HTTPServers, srv)
	}

	for _, svc := range sm.TCPServices {
		up := tcpReverseProxyUpstream{
			Name:    strings.Join([]string{svc.Namespace(), svc.Name()}, "__"),
			Servers: []tcpReverseProxyUpstreamServer{},
		}

		srv := tcpReverseProxyServer{
			ListenPort: svc.ListenPort,
			Upstream:   up.Name,
		}

		for _, ep := range svc.Endpoints {
			up.Servers = append(up.Servers, tcpReverseProxyUpstreamServer{
				Name: ep.Name,
				Host: ep.IP,
				Port: ep.Port,
			})
		}

		data.TCPUpstreams = append(data.TCPUpstreams, up)
		data.TCPServers = append(data.TCPServers, srv)
	}

	return &data
}

func renderConfig(cfg *NGINXConfig, sm *ServiceMap) ([]byte, error) {
	log.Printf("Rendering config")

	config := struct {
		ReverseProxyConfig *reverseProxyConfig
		*NGINXConfig
	}{
		ReverseProxyConfig: newReverseProxyConfig(cfg, sm),
		NGINXConfig:        cfg,
	}

	var buf bytes.Buffer
	if err := nginxTemplate.Execute(&buf, config); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func newLoggingNGINXManager() NGINXManager {
	return &loggingNGINXManager{status: nginxStatusStopped}
}

type loggingNGINXManager struct {
	status string
}

func (l *loggingNGINXManager) Status() (string, error) {
	log.Printf("called NGINXManager.Status()")
	return l.status, nil
}

func (l *loggingNGINXManager) Start() error {
	log.Printf("called NGINXManager.Start()")
	l.status = nginxStatusRunning
	return nil
}

func (l *loggingNGINXManager) Reload() error {
	log.Printf("called NGINXManager.Reload()")
	return nil
}

func (l *loggingNGINXManager) WriteConfig(sm *ServiceMap) error {
	log.Printf("called NGINXManager.WriteConfig(*ServiceMap) w/ %+v", sm)
	return nil
}
