package opentracing

import "context"

type contextKey struct{}

var activeSpanKey = contextKey{}

// ContextWithSpan 返回一个新的`context.Context`，它包含对span的引用。
// 如果span为空(nil)，将返回一个不包含活跃span的新context。
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	if span != nil {
		if tracerWithHook, ok := span.Tracer().(TracerContextWithSpanExtension); ok {
			ctx = tracerWithHook.ContextWithSpanHook(ctx, span)
		}
	}
	return context.WithValue(ctx, activeSpanKey, span)
}

// SpanFromContext 返回`ctx`中之前的`Span`(即上一个函数写进去的span)，如果没有找到`span`会返回`nil`
//
// 注意： context.Context != SpanContext: 前者是Go的进程内上下文传播机制，
// 后者包含有OpenTracing的Span识别和携带信息。
func SpanFromContext(ctx context.Context) Span {
	val := ctx.Value(activeSpanKey)
	if sp, ok := val.(Span); ok {
		return sp
	}
	return nil
}

// StartSpanFromContext 以`operationName`开始并返回一个Span，
// 使用在`ctx`中找到的 Span 作为新Span的`ChildOfRef`(即新span的父节点是ctx中的那个span)。
// 如果没有找到任何父级， StartSpanFromContext 将创建一个根(root)Span
//
// 第二个返回值是一个 context.Context 对象，包含有返回的 Span
//
// 样例:
//
//    SomeFunction(ctx context.Context, ...) {
//        sp, ctx := opentracing.StartSpanFromContext(ctx, "SomeFunction")
//        defer sp.Finish()
//        ...
//    }
func StartSpanFromContext(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, GlobalTracer(), operationName, opts...)
}

// StartSpanFromContextWithTracer 以`operationName`开始并返回一个Span，
// 使用在`ctx`中找到的 Span 作为新Span的`ChildOfRef`(即新span的父节点是ctx中的那个span)。
// 如果没有找到任何父级， StartSpanFromContext 将创建一个根(root)Span
//
// 它的行为与 StartSpanFromContext 相比，除了显示的tracer之外，其他是完全相同的。
// 对于 StartSpanFromContext, 它使用了 GlobalTracer。
func StartSpanFromContextWithTracer(ctx context.Context, tracer Tracer, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, ChildOf(parentSpan.Context()))
	}
	span := tracer.StartSpan(operationName, opts...)
	return span, ContextWithSpan(ctx, span)
}
