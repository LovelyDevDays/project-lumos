package client

type clientOptions struct {
	host string
	port string
}

var defaultClientOptions = clientOptions{
	host: "passage-retrieval-service",
	port: "50051",
}

type Option func(*clientOptions)

func WithHost(host string) Option {
	return func(opt *clientOptions) {
		opt.host = host
	}
}

func WithPort(port string) Option {
	return func(opt *clientOptions) {
		opt.port = port
	}
}
