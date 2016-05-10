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

events {
    worker_connections 512;
}

http {
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

func newNGINXManager() *NGINXManager {
	return &NGINXManager{cfg: DefaultNGINXConfig}
}

type NGINXConfig struct {
	ConfigFile string
	PIDFile    string
}

type NGINXManager struct {
	cfg NGINXConfig
}

func (n *NGINXManager) Status() (string, error) {
	if _, err := os.Stat(n.cfg.PIDFile); err != nil {
		if os.IsNotExist(err) {
			return nginxStatusStopped, nil
		} else {
			return nginxStatusUnknown, err
		}
	}

	return nginxStatusRunning, nil
}

func (n *NGINXManager) WriteConfig(sm *ServiceMap) error {
	cfg, err := renderConfig(&n.cfg, sm)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(n.cfg.ConfigFile, cfg, os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

func (n *NGINXManager) assertConfigOK() error {
	return n.run("-t")
}

func (n *NGINXManager) Start() error {
	return n.run()
}

func (n *NGINXManager) Reload() error {
	return n.run("-s", "reload")
}

func (n *NGINXManager) run(args ...string) error {
	args = append([]string{"-c", n.cfg.ConfigFile}, args...)
	output, err := exec.Command("nginx", args...).CombinedOutput()
	if err != nil {
		log.Printf("nginx command failed w/ output:\n%s", output)
		return err
	}
	return nil
}

func renderConfig(cfg *NGINXConfig, sm *ServiceMap) ([]byte, error) {
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
