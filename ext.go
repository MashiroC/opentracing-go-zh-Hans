package opentracing

import (
	"context"
)

// TracerContextWithSpanExtension 是一个扩展接口，Tracer的实现可能要实现该接口。
// 它允许Tracer在执行 ContextWithSpan 时控制go的context。
//
// 此扩展的主要目的是从 opentracing API 到其他一些跟踪 API 的适配器。
type TracerContextWithSpanExtension interface {
	// ContextWithSpanHook 当 Tracer 实现了此接口时，可以被 ContextWithSpan 调用。
	// 它允许调用者放置一些额外的信息在context中，并时 ContextWithSpan 函数对调用者可用。
	//
	// 该hook在 ContextWithSpan 函数会在span放在context前执行。
	ContextWithSpanHook(ctx context.Context, span Span) context.Context
}
