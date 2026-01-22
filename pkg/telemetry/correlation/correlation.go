package correlation

import (
	"context"

	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel/trace"
)

// correlationKey is an unexported type for context keys within this package.
type correlationKey struct{}

// ExtractCorrelationID fetches a correlation ID from the context if present.
func ExtractCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(correlationKey{}).(string); ok {
		return val
	}
	return ""
}

// ContextWithCorrelationID sets the correlation ID onto the context.
func ContextWithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationKey{}, id)
}

// InjectCorrelationID is kept for backwards compatibility and delegates to ContextWithCorrelationID.
func InjectCorrelationID(ctx context.Context, id string) context.Context {
	return ContextWithCorrelationID(ctx, id)
}

// EnsureCorrelationID guarantees a correlation ID on the context, generating one when missing.
func EnsureCorrelationID(ctx context.Context) (context.Context, string) {
	cid := ExtractCorrelationID(ctx)
	if cid == "" {
		cid = ulid.Make().String()
	}
	return ContextWithCorrelationID(ctx, cid), cid
}

// ContextWithRemoteSpan seeds the context with a remote span if valid identifiers are provided.
func ContextWithRemoteSpan(ctx context.Context, traceIDHex, spanIDHex string) context.Context {
	if traceIDHex == "" || spanIDHex == "" {
		return ctx
	}

	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		return ctx
	}
	spanID, err := trace.SpanIDFromHex(spanIDHex)
	if err != nil {
		return ctx
	}

	parent := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true})
	return trace.ContextWithSpanContext(ctx, parent)
}
