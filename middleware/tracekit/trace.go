package tracekit

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"strings"

	"github.com/clubpay/ronykit"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const (
	w3cTraceParent = "traceparent"
	w3cState       = "tracestate"
	b3Single       = "b3"
	b3TraceID      = "x-b3-traceid"
	b3ParentSpanID = "x-b3-parentspanid"
	b3SpanID       = "x-b3-spanid"
	b3Sampled      = "x-b3-sampled"
	b3Flags        = "x-b3-flags"
)

type TracePropagator int

const (
	w3cPropagator TracePropagator = iota
	b3Propagator
)

func B3(name string, opts ...Option) ronykit.HandlerFunc {
	cfg := &config{
		tracerName: name,
		propagator: b3Propagator,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return withTracer(cfg)
}

func W3C(name string, opts ...Option) ronykit.HandlerFunc {
	cfg := &config{
		tracerName: name,
		propagator: w3cPropagator,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return withTracer(cfg)
}

func withTracer(cfg *config) ronykit.HandlerFunc {
	var (
		traceCtx     propagation.TextMapPropagator
		traceCarrier func(ctx *ronykit.Context) propagation.TextMapCarrier
	)

	switch cfg.propagator {
	case b3Propagator:
		traceCtx = b3.New(b3.WithInjectEncoding(b3.B3SingleHeader))
		traceCarrier = newB3Carrier
	default:
		traceCtx = propagation.TraceContext{}
		traceCarrier = newW3CCarrier
	}

	var (
		spanOpts []trace.SpanStartOption
		kvs      []attribute.KeyValue
	)

	if cfg.serviceName != "" {
		kvs = append(kvs, semconv.ServiceNameKey.String(cfg.serviceName))
	}
	if cfg.env != "" {
		kvs = append(kvs, semconv.DeploymentEnvironmentKey.String(cfg.env))
	}

	if len(kvs) > 0 {
		spanOpts = append(spanOpts, trace.WithAttributes(kvs...))
	}

	return func(ctx *ronykit.Context) {
		rc, ok := ctx.Conn().(ronykit.RESTConn)
		if ok {
			spanOpts = append(spanOpts,
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(rc.GetMethod()),
					semconv.HTTPStatusCodeKey.Int(ctx.GetStatusCode()),
				),
			)
		}

		userCtx, span := otel.Tracer(cfg.tracerName).
			Start(
				traceCtx.Extract(ctx.Context(), traceCarrier(ctx)),
				ctx.Route(),
				spanOpts...,
			)
		ctx.SetUserContext(userCtx)
		ctx.Next()
		span.End()
	}
}

type w3cCarrier struct {
	traceParent string
	traceState  string
	ctx         *ronykit.Context
}

func newW3CCarrier(ctx *ronykit.Context) propagation.TextMapCarrier {
	c := w3cCarrier{ctx: ctx}
	c.ctx.Conn().Walk(
		func(key string, v string) bool {
			if strings.EqualFold(w3cTraceParent, key) {
				c.traceParent = v
			} else if strings.EqualFold(w3cState, key) {
				c.traceState = v
			}

			return true
		},
	)

	return c
}

func (c w3cCarrier) Get(key string) string {
	switch key {
	case w3cTraceParent:
		return c.traceParent
	case w3cState:
		return c.traceState
	}

	v, ok := c.ctx.Get(key).(string)
	if !ok {
		return ""
	}

	return v
}

func (c w3cCarrier) Set(key string, value string) {
	c.ctx.Set(key, value)
}

func (c w3cCarrier) Keys() []string {
	var keys []string
	c.ctx.Conn().Walk(
		func(key string, _ string) bool {
			keys = append(keys, key)

			return true
		},
	)

	return keys
}

type b3Carrier struct {
	b3           string
	traceID      string
	parentSpanID string
	spanID       string
	sampled      string
	flags        string
	ctx          *ronykit.Context
}

func newB3Carrier(ctx *ronykit.Context) propagation.TextMapCarrier {
	b3.New()
	c := b3Carrier{ctx: ctx}
	c.ctx.Conn().Walk(
		func(key string, v string) bool {
			switch {
			case strings.EqualFold(b3Single, key):
				c.b3 = v
			case strings.EqualFold(b3TraceID, key):
				c.traceID = v
			case strings.EqualFold(b3SpanID, key):
				c.spanID = v
			case strings.EqualFold(b3ParentSpanID, key):
				c.parentSpanID = v
			case strings.EqualFold(b3Sampled, key):
				c.sampled = v
			case strings.EqualFold(b3Flags, key):
				c.flags = v
			}

			return true
		},
	)

	return c
}

func (c b3Carrier) Get(key string) string {
	switch key {
	case b3Single:
		return c.b3
	case b3TraceID:
		return c.traceID
	case b3SpanID:
		return c.spanID
	case b3ParentSpanID:
		return c.parentSpanID
	case b3Sampled:
		return c.sampled
	case b3Flags:
		return c.flags
	}

	v, ok := c.ctx.Get(key).(string)
	if !ok {
		return ""
	}

	return v
}

func (c b3Carrier) Set(key string, value string) {
	c.ctx.Set(key, value)
}

func (c b3Carrier) Keys() []string {
	var keys []string
	c.ctx.Conn().Walk(
		func(key string, _ string) bool {
			keys = append(keys, key)

			return true
		},
	)

	return keys
}
