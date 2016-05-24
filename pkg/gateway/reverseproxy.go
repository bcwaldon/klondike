package gateway

type reverseProxyConfig struct {
	HTTPServers   []httpReverseProxyServer
	HTTPUpstreams []httpReverseProxyUpstream
	TCPServers    []tcpReverseProxyServer
	TCPUpstreams  []tcpReverseProxyUpstream
}

type httpReverseProxyServer struct {
	Name          string
	AltNames      []string
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
	Servers []httpReverseProxyUpstreamServer
}

type httpReverseProxyUpstreamServer struct {
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
	Servers []tcpReverseProxyUpstreamServer
}

type tcpReverseProxyUpstreamServer struct {
	Name string
	Host string
	Port int
}
