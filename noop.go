package opentracing

import "github.com/opentracing/opentracing-go/log"

// NoopTracer 是一个微不足道的，最小开销的 Tracer 实现，该实现中所有方法都是空操作(no-ops)。
//
// 该实现主要用于库，比如RPC框架，它使得链路追踪成为了一个用户可选的功能。
// 一个空操作(no-op)的实现允许类库用它作为默认的 Tracer，不需要检查tracer的实例是否为空(nil)
// （翻译注：就是指你可以把global tracer设置成这个，这样的话在代码里就不用每次调用tracer都检查有没有开启或者tracer是不是空了）
//
// 基于同样的原因，NoopTracer 是全局tracer的默认值。
// （见 GlobalTracer 和 SetGlobalTracer）
//
// 警告(WARNING)： NoopTracer 没有支持携带数据(baggage)的传播
type NoopTracer struct{}

type noopSpan struct{}
type noopSpanContext struct{}

var (
	defaultNoopSpanContext SpanContext = noopSpanContext{}
	defaultNoopSpan        Span        = noopSpan{}
	defaultNoopTracer      Tracer      = NoopTracer{}
)

const (
	emptyString = ""
)

// noopSpanContext:
func (n noopSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

// noopSpan:
func (n noopSpan) Context() SpanContext                                  { return defaultNoopSpanContext }
func (n noopSpan) SetBaggageItem(key, val string) Span                   { return n }
func (n noopSpan) BaggageItem(key string) string                         { return emptyString }
func (n noopSpan) SetTag(key string, value interface{}) Span             { return n }
func (n noopSpan) LogFields(fields ...log.Field)                         {}
func (n noopSpan) LogKV(keyVals ...interface{})                          {}
func (n noopSpan) Finish()                                               {}
func (n noopSpan) FinishWithOptions(opts FinishOptions)                  {}
func (n noopSpan) SetOperationName(operationName string) Span            { return n }
func (n noopSpan) Tracer() Tracer                                        { return defaultNoopTracer }
func (n noopSpan) LogEvent(event string)                                 {}
func (n noopSpan) LogEventWithPayload(event string, payload interface{}) {}
func (n noopSpan) Log(data LogData)                                      {}

// StartSpan 实现 Tracer 接口
func (n NoopTracer) StartSpan(operationName string, opts ...StartSpanOption) Span {
	return defaultNoopSpan
}

// Inject 实现 Tracer 接口
func (n NoopTracer) Inject(sp SpanContext, format interface{}, carrier interface{}) error {
	return nil
}

// Extract 实现 Tracer 接口
func (n NoopTracer) Extract(format interface{}, carrier interface{}) (SpanContext, error) {
	return nil, ErrSpanContextNotFound
}
