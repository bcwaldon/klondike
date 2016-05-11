package gateway

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
    server {
        listen 7332;
        location /health {
            return 200 'Healthy!';
        }
    }
{{ range $svc := .ServiceMap.Services }}
    server {
        listen {{ $svc.ListenPort }};
        location / {
            proxy_pass http://{{ $svc.Namespace }}__{{ $svc.Name }};
        }
    }
    upstream {{ $svc.Namespace }}__{{ $svc.Name }} {
{{ range $ep := $svc.Endpoints }}
        server {{ $ep.IP }};  # {{ $ep.Name }}
{{- end }}
    }
{{ end }}
}
`

	nginxTemplate = template.Must(template.New("nginx").Parse(nginxTemplateData))

	DefaultNGINXConfig = NGINXConfig{
		ConfigFile: "/etc/nginx/nginx.conf",
		PIDFile:    "/var/run/nginx.pid",
	}
)

const (
	nginxStatusRunning = "running"
	nginxStatusStopped = "stopped"
	nginxStatusUnknown = "unknown"
)

type NGINXConfig struct {
	ConfigFile string
	PIDFile    string
}

type NGINXManager interface {
	Status() (string, error)
	WriteConfig(*ServiceMap) error
	Start() error
	Reload() error
}

func newNGINXManager() NGINXManager {
	return &nginxManager{cfg: DefaultNGINXConfig}
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
	output, err := exec.Command("nginx", args...).CombinedOutput()
	if err != nil {
		log.Printf("nginx command failed w/ output:\n%s", output)
		return err
	} else {
		log.Printf("nginx command success w/ output:\n%s", output)
	}
	return nil
}

func renderConfig(cfg *NGINXConfig, sm *ServiceMap) ([]byte, error) {
	log.Printf("Rendering config")
	config := struct {
		*NGINXConfig
		*ServiceMap
	}{
		NGINXConfig: cfg,
		ServiceMap:  sm,
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
