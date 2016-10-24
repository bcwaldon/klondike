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

type ReverseProxyConfigGetter interface {
	ReverseProxyConfig() (*reverseProxyConfig, error)
}

type reverseProxyConfig struct {
	HTTPServers   []httpReverseProxyServer
	HTTPUpstreams []httpReverseProxyUpstream
	TCPServers    []tcpReverseProxyServer
	TCPUpstreams  []tcpReverseProxyUpstream
}

type httpReverseProxyServer struct {
	Name          string
	AltNames      []string
	DefaultServer bool
	ListenPort    int
	Locations     []httpReverseProxyLocation
	StaticCode    int
	StaticMessage string
}

type httpReverseProxyLocation struct {
	Path          string
	StaticCode    int
	StaticMessage string
	Upstream      string
}

type httpReverseProxyUpstream struct {
	Name    string
	Servers []reverseProxyUpstreamServer
}

type reverseProxyUpstreamServer struct {
	Name string
	Host string
	Port int
}

type tcpReverseProxyServer struct {
	ListenPort int
	Upstream   string
}

type tcpReverseProxyUpstream struct {
	Name    string
	Servers []reverseProxyUpstreamServer
}
