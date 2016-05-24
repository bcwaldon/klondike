package gateway

import (
	"strings"
	"testing"
)

type fakeReverseProxyConfigGetter struct {
	rc  reverseProxyConfig
	err error
}

func (fsm *fakeReverseProxyConfigGetter) ReverseProxyConfig() (*reverseProxyConfig, error) {
	return &fsm.rc, fsm.err
}

func TestRenderHTTPServers(t *testing.T) {
	fsm := fakeReverseProxyConfigGetter{
		rc: reverseProxyConfig{
			HTTPServers: []httpReverseProxyServer{
				httpReverseProxyServer{
					ListenPort: 9001,
					StaticCode: 202,
				},
			},
			HTTPUpstreams: []httpReverseProxyUpstream{},
		},
	}

	cfg := DefaultNGINXConfig
	got, err := renderConfig(&cfg, &fsm.rc)
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
    server_names_hash_bucket_size 128;

    server {
        listen 9001;
        
        return 202;
    }


}


stream {



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
