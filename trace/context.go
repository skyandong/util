package trace

import "context"

// FullTraceSpanKeyType in context
type FullTraceSpanKeyType string

// FullTraceSpanKey for span in context
const FullTraceSpanKey FullTraceSpanKeyType = "full-trace"

// SpanToContext add span to context
func SpanToContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, FullTraceSpanKey, span)
}

// SpanFromContext extract span from context
func SpanFromContext(ctx context.Context) *Span {
	if v := ctx.Value(FullTraceSpanKey); v != nil {
		if s, ok := v.(*Span); ok {
			return s
		}
	}
	// be compatible with gin.Context
	if v := ctx.Value(string(FullTraceSpanKey)); v != nil {
		if s, ok := v.(*Span); ok {
			return s
		}
	}
	return nil
}
