package server

type serverOptions struct {
	port string

	serviceV1 ServiceV1
}

var defaultServerOptions = serverOptions{
	port: "50051",
}

type Option func(*serverOptions)

func WithPort(port string) Option {
	return func(opt *serverOptions) {
		opt.port = port
	}
}

func WithServiceV1(service ServiceV1) Option {
	return func(opt *serverOptions) {
		opt.serviceV1 = service
	}
}
