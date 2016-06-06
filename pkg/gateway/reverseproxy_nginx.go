package gateway

import (
	"bytes"
	"github.com/bcwaldon/farva/pkg/logger"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var (
	nginxTemplateData = `
pid {{ .NGINXConfig.PIDFile }};
error_log {{ .NGINXConfig.ErrorLog }};
daemon on;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log {{ .NGINXConfig.AccessLog }} main;
{{ range $srv := $.ReverseProxyConfig.HTTPServers }}
    server {
        listen {{ $srv.ListenPort }}{{ if $srv.DefaultServer }} default_server{{ end }};
        {{ if $srv.Name }}server_name {{ $srv.Name }}{{ if $srv.AltNames }} {{ join $srv.AltNames " " }}{{ end }};{{ end }}
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
{{- end }}
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

	nginxTemplate = template.Must(template.New("nginx").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(nginxTemplateData))

	DefaultNGINXConfig = NGINXConfig{
		ClusterZone: "example.com",
		ConfigFile:  "/etc/nginx/nginx.conf",
		PIDFile:     "/var/run/nginx.pid",
		HealthPort:  7332,
		AccessLog:   "/dev/stdout",
		ErrorLog:    "/dev/stderr",
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
	ErrorLog    string
	AccessLog   string
}

func newNGINXConfig(hp int, cz string, errorLog string, accessLog string) NGINXConfig {
	cfg := DefaultNGINXConfig
	cfg.HealthPort = hp
	cfg.ClusterZone = cz
	cfg.ErrorLog = errorLog
	cfg.AccessLog = accessLog
	return cfg
}

type NGINXManager interface {
	Status() (string, error)
	WriteConfig(*reverseProxyConfig) error
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
	logger.Log.Info("Checking status")
	if _, err := os.Stat(n.cfg.PIDFile); err != nil {
		if os.IsNotExist(err) {
			return nginxStatusStopped, nil
		} else {
			return nginxStatusUnknown, err
		}
	}

	return nginxStatusRunning, nil
}

func (n *nginxManager) WriteConfig(rc *reverseProxyConfig) error {
	cfg, err := renderConfig(&n.cfg, rc)
	logger.Log.Infof("About to write config: %s", string(cfg))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(n.cfg.ConfigFile, cfg, os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

func (n *nginxManager) assertConfigOK() error {
	_, err := n.runCombinedOutput("-t")
	return err
}

func (n *nginxManager) Start() error {
	if err := n.assertConfigOK(); err != nil {
		logger.Log.Info("Configuration is invalid, aborting start")
		return err
	}
	logger.Log.Info("Starting nginx")
	return n.run()
}

func (n *nginxManager) Reload() error {
	if err := n.assertConfigOK(); err != nil {
		return err
	}
	logger.Log.Info("Reloading nginx")
	return n.run("-s", "reload")
}

func (n *nginxManager) run(args ...string) error {
	args = append([]string{"-c", n.cfg.ConfigFile}, args...)
	logger.Log.Infof("Calling run on nginx with args: %q", args)
	err := exec.Command("nginx", args...).Run()
	if err != nil {
		logger.Log.Infof("nginx command failed w/ err: %v", err)
		return err
	} else {
		logger.Log.Info("nginx command success")
	}
	return nil
}

func (n *nginxManager) runCombinedOutput(args ...string) (string, error) {
	args = append([]string{"-c", n.cfg.ConfigFile}, args...)
	logger.Log.Infof("Calling run on nginx with args: %q", args)
	output, err := exec.Command("nginx", args...).CombinedOutput()
	if err != nil {
		logger.Log.Infof("nginx command failed w/ err: %v, output:%s", err, output)
		return "", err
	} else {
		logger.Log.Info("nginx command success")
	}
	return string(output), nil
}

func renderConfig(cfg *NGINXConfig, rc *reverseProxyConfig) ([]byte, error) {
	logger.Log.Info("Rendering config")

	config := struct {
		ReverseProxyConfig *reverseProxyConfig
		*NGINXConfig
	}{
		ReverseProxyConfig: rc,
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
	logger.Log.Info("called NGINXManager.Status()")
	return l.status, nil
}

func (l *loggingNGINXManager) Start() error {
	logger.Log.Info("called NGINXManager.Start()")
	l.status = nginxStatusRunning
	return nil
}

func (l *loggingNGINXManager) Reload() error {
	logger.Log.Info("called NGINXManager.Reload()")
	return nil
}

func (l *loggingNGINXManager) WriteConfig(rc *reverseProxyConfig) error {
	logger.Log.Infof("called NGINXManager.WriteConfig(*reverseProxyConfig) w/ %+v", rc)
	return nil
}
