package tracekit

import (
	"github.com/clubpay/ronykit"
	"go.opentelemetry.io/otel/trace"
)

func Span(ctx *ronykit.Context) trace.Span {
	return trace.SpanFromContext(ctx.Context())
}
