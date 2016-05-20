package gateway

import (
	"strings"
	"testing"
)

type fakeServiceMapper struct {
	sm  ServiceMap
	err error
}

func (fsm *fakeServiceMapper) ServiceMap() (*ServiceMap, error) {
	return &fsm.sm, fsm.err
}

func TestRender(t *testing.T) {
	fsm := fakeServiceMapper{
		sm: ServiceMap{
			ServiceGroups: []ServiceGroup{
				ServiceGroup{
					Name:      "ing1",
					Namespace: "svc1",
					Services: []Service{
						Service{
							Namespace: "ns1",
							Name:      "svc1",
							Endpoints: []Endpoint{
								Endpoint{Name: "pod1", IP: "10.0.0.1"},
								Endpoint{Name: "pod2", IP: "10.0.0.2"},
								Endpoint{Name: "pod3", IP: "10.0.0.3"},
							},
						},
						Service{
							Namespace: "ns1",
							Name:      "svc2",
							Path:      "/v0",
							Endpoints: []Endpoint{
								Endpoint{Name: "pod1", IP: "10.0.0.4"},
								Endpoint{Name: "pod2", IP: "10.0.0.5"},
								Endpoint{Name: "pod3", IP: "10.0.0.6"},
							},
						},
					},
				},
			},
		},
	}

	cfg := DefaultNGINXConfig
	got, err := renderConfig(&cfg, &fsm.sm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `
pid /var/run/nginx.pid;
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

    server {
        listen 7331;
        server_name ing1.svc1.example.com;

        location / {
            proxy_pass http://ns1__ing1__svc1;
        }
        location /v0 {
            proxy_pass http://ns1__ing1__svc2;
        }
    }

    upstream ns1__ing1__svc1 {

        server 10.0.0.1:0;  # pod1
        server 10.0.0.2:0;  # pod2
        server 10.0.0.3:0;  # pod3
    }
    upstream ns1__ing1__svc2 {

        server 10.0.0.4:0;  # pod1
        server 10.0.0.5:0;  # pod2
        server 10.0.0.6:0;  # pod3
    }

}
`
	if want != string(got) {
		t.Fatalf(
			"unexpected output: want=%sgot=%s",
			strings.Replace(want, " ", "÷", -1),
			strings.Replace(string(got), " ", "÷", -1),
		)
	}
}
