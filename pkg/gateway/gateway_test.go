package gateway

import (
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
			Services: []Service{
				Service{
					Namespace:  "ns1",
					Name:       "svc1",
					ListenPort: 30001,
					Endpoints: []Endpoint{
						Endpoint{Name: "pod1", IP: "10.0.0.1"},
						Endpoint{Name: "pod2", IP: "10.0.0.2"},
						Endpoint{Name: "pod3", IP: "10.0.0.3"},
					},
				},
			},
		},
	}

	gw := Gateway{sm: &fsm}
	got, err := gw.Render()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `
http {

    server ns1__svc1 {
        listen 30001;
    }
    upstream ns1__svc1 {

        server 10.0.0.1;  # pod1
        server 10.0.0.2;  # pod2
        server 10.0.0.3;  # pod3
    }

}
`
	if want != string(got) {
		t.Fatalf("unexpected output: want=\n%s\ngot=%s", want, string(got))
	}
}
