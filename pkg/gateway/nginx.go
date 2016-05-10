package gateway

import (
	"bytes"
	"text/template"
)

var (
	nginxTemplateData = `
http {
{{ range $svc := .ServiceMap.Services }}
    server {{ $svc.Namespace }}__{{ $svc.Name }} {
        listen {{ $svc.ListenPort }};
    }
    upstream {{ $svc.Namespace }}__{{ $svc.Name }} {
    {{ range $name, $ep := $svc.Endpoints }}
        server {{ $ep }};  # {{ $name }}
    {{- end }}
    }
{{ end }}
}
`

	nginxTemplate = template.Must(template.New("nginx").Parse(nginxTemplateData))
)

func renderNginxConfig(sm *ServiceMap) ([]byte, error) {
	config := struct {
		*ServiceMap
	}{
		ServiceMap: sm,
	}

	var buf bytes.Buffer
	if err := nginxTemplate.Execute(&buf, config); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
