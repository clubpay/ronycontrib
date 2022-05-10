package tracekit

type Option func(cfg *config)

type config struct {
	tracerName  string
	propagator  TracePropagator
	serviceName string
	env         string
}

func ServiceName(name string) Option {
	return func(cfg *config) {
		cfg.serviceName = name
	}
}

func Env(env string) Option {
	return func(cfg *config) {
		cfg.env = env
	}
}
