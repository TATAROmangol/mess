package requestmeta

type HttpMetadata struct {
	Method    string `json:"method"`
	URL       string `json:"url"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent,omitempty"`
}

type GrpcMetadata struct {
	Method   string `json:"method"`
	PeerAddr string `json:"peer_addr,omitempty"`
}
