package opentracing

type registeredTracer struct {
	tracer       Tracer
	isRegistered bool
}

var (
	globalTracer = registeredTracer{NoopTracer{}, false}
)

// SetGlobalTracer 设置一个[单例]的追踪系统。 Tracer 可以使用 GlobalTracer() 返回。
// 不管谁使用 GlobalTracer（而不是指直接管理 opentracing.Tracer 的实例），
// 都应该在main()中尽早的调用 SetGlobalTracer，应在`StartSpan`的调用之前。
// 在调用`SetGlobalTracer`之前，任何通过`StartSpan`创建的Span都是来自noop的。
func SetGlobalTracer(tracer Tracer) {
	globalTracer = registeredTracer{tracer, true}
}

// GloablTracer 返回`Tracer`实现的全局单例。
// 在调用`SetGlobalTracer()`之前，`GlobalTracer()`返回的是noop实现，它会丢掉所有的数据。
func GlobalTracer() Tracer {
	return globalTracer.tracer
}

// StartSpan 遵从 Tracer.StartSpan，见 `GlobalTracer()`。
func StartSpan(operationName string, opts ...StartSpanOption) Span {
	return globalTracer.tracer.StartSpan(operationName, opts...)
}

// InitGlobalTracer 已废弃(deprecated)，请使用 SetGlobalTracer。
// InitGlobalTracer is deprecated. Please use SetGlobalTracer.
func InitGlobalTracer(tracer Tracer) {
	SetGlobalTracer(tracer)
}

// IsGlobalTracerRegistered 返回一个布尔值去判断tracer是否已经在全局注册
// IsGlobalTracerRegistered returns a `bool` to indicate if a tracer has been globally registered
func IsGlobalTracerRegistered() bool {
	return globalTracer.isRegistered
}
