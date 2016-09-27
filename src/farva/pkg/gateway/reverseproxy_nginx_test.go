/*
Copyright 2016 Planet Labs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

func TestRenderConfig(t *testing.T) {
	tests := []struct {
		rc   reverseProxyConfig
		want string
	}{
		// static code with and without message, no location
		{
			rc: reverseProxyConfig{
				HTTPServers: []httpReverseProxyServer{
					httpReverseProxyServer{
						ListenPort: 9001,
						StaticCode: 202,
					},
					httpReverseProxyServer{
						ListenPort:    9002,
						StaticCode:    203,
						StaticMessage: "ping pong",
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;


    server {
        listen 9001;
        
        return 202;
    }

    server {
        listen 9002;
        
        return 203 'ping pong';
    }



    server {
        listen 9001;
        server_name localhost;

        access_log off;
        allow 127.0.0.1;
        deny all;

        location /nginx_status {
          stub_status on;
        }
    }




}

stream {


}
`,
		},

		// single location, static code w/ and w/o message
		{
			rc: reverseProxyConfig{
				HTTPServers: []httpReverseProxyServer{
					httpReverseProxyServer{
						ListenPort: 9001,
						Locations: []httpReverseProxyLocation{
							httpReverseProxyLocation{
								Path:       "/foo",
								StaticCode: 202,
							},
						},
					},
					httpReverseProxyServer{
						ListenPort: 9002,
						Locations: []httpReverseProxyLocation{
							httpReverseProxyLocation{
								Path:          "/bar/baz",
								StaticCode:    203,
								StaticMessage: "ping pong",
							},
						},
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;


    server {
        listen 9001;
        
        
        location /foo {
            return 202;
        }

    }

    server {
        listen 9002;
        
        
        location /bar/baz {
            return 203 'ping pong';
        }

    }



    server {
        listen 9001;
        server_name localhost;

        access_log off;
        allow 127.0.0.1;
        deny all;

        location /nginx_status {
          stub_status on;
        }
    }




}

stream {


}
`,
		},

		// one HTTP server, multiple upstreams and paths
		{
			rc: reverseProxyConfig{
				HTTPServers: []httpReverseProxyServer{
					httpReverseProxyServer{
						ListenPort: 9001,
						Locations: []httpReverseProxyLocation{
							httpReverseProxyLocation{
								Path:     "/abc",
								Upstream: "foo",
							},
							httpReverseProxyLocation{
								Path:     "/def",
								Upstream: "bar",
							},
						},
					},
				},
				HTTPUpstreams: []httpReverseProxyUpstream{
					httpReverseProxyUpstream{
						Name: "foo",
						Servers: []reverseProxyUpstreamServer{
							reverseProxyUpstreamServer{
								Name: "ping",
								Host: "ping.example.com",
								Port: 443,
							},
							reverseProxyUpstreamServer{
								Name: "pong",
								Host: "pong.example.com",
								Port: 80,
							},
						},
					},
					httpReverseProxyUpstream{
						Name: "bar",
						Servers: []reverseProxyUpstreamServer{
							reverseProxyUpstreamServer{
								Name: "ding",
								Host: "ding.example.com",
								Port: 443,
							},
							reverseProxyUpstreamServer{
								Name: "dong",
								Host: "dong.example.com",
								Port: 80,
							},
						},
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;


    server {
        listen 9001;
        
        
        location /abc {
            
            proxy_pass http://foo;
        }

        location /def {
            
            proxy_pass http://bar;
        }

    }



    server {
        listen 9001;
        server_name localhost;

        access_log off;
        allow 127.0.0.1;
        deny all;

        location /nginx_status {
          stub_status on;
        }
    }



    upstream foo {

        server ping.example.com:443;  # ping
        server pong.example.com:80;  # pong
        keepalive 64;
    }


    upstream bar {

        server ding.example.com:443;  # ding
        server dong.example.com:80;  # dong
        keepalive 64;
    }

}

stream {


}
`,
		},

		// two TCP servers, various upstreams
		{
			rc: reverseProxyConfig{
				TCPServers: []tcpReverseProxyServer{
					tcpReverseProxyServer{
						ListenPort: 9001,
						Upstream:   "foo",
					},
					tcpReverseProxyServer{
						ListenPort: 9002,
						Upstream:   "bar",
					},
				},
				TCPUpstreams: []tcpReverseProxyUpstream{
					tcpReverseProxyUpstream{
						Name: "foo",
						Servers: []reverseProxyUpstreamServer{
							reverseProxyUpstreamServer{
								Name: "ping",
								Host: "ping.example.com",
								Port: 443,
							},
						},
					},
					tcpReverseProxyUpstream{
						Name: "bar",
						Servers: []reverseProxyUpstreamServer{
							reverseProxyUpstreamServer{
								Name: "ding",
								Host: "ding.example.com",
								Port: 443,
							},
							reverseProxyUpstreamServer{
								Name: "dong",
								Host: "dong.example.com",
								Port: 80,
							},
						},
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;




}

stream {

    server {
        listen 9001;
        proxy_pass foo;
    }

    server {
        listen 9002;
        proxy_pass bar;
    }



    upstream foo {

        server ping.example.com:443;  # ping
    }


    upstream bar {

        server ding.example.com:443;  # ding
        server dong.example.com:80;  # dong
    }

}
`,
		},
		// Multiple AltNames
		{
			rc: reverseProxyConfig{
				HTTPServers: []httpReverseProxyServer{
					httpReverseProxyServer{
						Name:       "default.example.com",
						AltNames:   []string{"test.example.com", "foo.bar.com"},
						ListenPort: 9001,
						StaticCode: 202,
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;


    server {
        listen 9001;
        server_name default.example.com test.example.com foo.bar.com;
        return 202;
    }



    server {
        listen 9001;
        server_name localhost;

        access_log off;
        allow 127.0.0.1;
        deny all;

        location /nginx_status {
          stub_status on;
        }
    }


}

stream {


}
`,
		},
		// DefaultServer
		{
			rc: reverseProxyConfig{
				HTTPServers: []httpReverseProxyServer{
					httpReverseProxyServer{
						ListenPort:    9001,
						DefaultServer: true,
						StaticCode:    202,
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;


    server {
        listen 9001 default_server;
        
        return 202;
    }



    server {
        listen 9001;
        server_name localhost;

        access_log off;
        allow 127.0.0.1;
        deny all;

        location /nginx_status {
          stub_status on;
        }
    }


}

stream {


}
`,
		},
		// A Server with no upstreams.
		{
			rc: reverseProxyConfig{
				HTTPUpstreams: []httpReverseProxyUpstream{
					httpReverseProxyUpstream{
						Name:    "foo",
						Servers: []reverseProxyUpstreamServer{},
					},
				},
				TCPUpstreams: []tcpReverseProxyUpstream{
					tcpReverseProxyUpstream{
						Name:    "foo",
						Servers: []reverseProxyUpstreamServer{},
					},
				},
			},
			want: `
pid /var/run/nginx.pid;
error_log /dev/stderr;
daemon on;
worker_processes auto;

events {
    worker_connections 512;
}

http {
    server_names_hash_bucket_size 128;
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /dev/stdout main;

    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_set_header Host $host;






}

stream {




}
`,
		},
	}

	for i, tt := range tests {
		fsm := fakeReverseProxyConfigGetter{rc: tt.rc}
		cfg := DefaultNGINXConfig
		got, err := renderConfig(&cfg, &fsm.rc)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		}

		if tt.want != string(got) {
			wantPretty := strings.Replace(tt.want, " ", "÷", -1)
			gotPretty := strings.Replace(string(got), " ", "÷", -1)
			t.Errorf("case %d: unexpected output: want=%sgot=%s", i, wantPretty, gotPretty)
		}
	}
}
