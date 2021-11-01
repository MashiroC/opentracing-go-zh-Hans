package opentracing

import (
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

const testHTTPHeaderPrefix = "testprefix-"

// testTracer 是一个大多数操作是空操作的 Tracer 实现。
// 它使得单元测试验证某些函数有没有被调用成为可能。
type testTracer struct{}

var fakeIDSource = 1

func nextFakeID() int {
	fakeIDSource++
	return fakeIDSource
}

type testSpanContext struct {
	HasParent bool
	FakeID    int
}

func (n testSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

type testSpan struct {
	spanContext   testSpanContext
	OperationName string
	StartTime     time.Time
	Tags          map[string]interface{}
}

func (n testSpan) Equal(os Span) bool {
	other, ok := os.(testSpan)
	if !ok {
		return false
	}
	if n.spanContext != other.spanContext {
		return false
	}
	if n.OperationName != other.OperationName {
		return false
	}
	if !n.StartTime.Equal(other.StartTime) {
		return false
	}
	if len(n.Tags) != len(other.Tags) {
		return false
	}

	for k, v := range n.Tags {
		if ov, ok := other.Tags[k]; !ok || ov != v {
			return false
		}
	}

	return true
}

// testSpan:
func (n testSpan) Context() SpanContext                                  { return n.spanContext }
func (n testSpan) SetTag(key string, value interface{}) Span             { return n }
func (n testSpan) Finish()                                               {}
func (n testSpan) FinishWithOptions(opts FinishOptions)                  {}
func (n testSpan) LogFields(fields ...log.Field)                         {}
func (n testSpan) LogKV(kvs ...interface{})                              {}
func (n testSpan) SetOperationName(operationName string) Span            { return n }
func (n testSpan) Tracer() Tracer                                        { return testTracer{} }
func (n testSpan) SetBaggageItem(key, val string) Span                   { return n }
func (n testSpan) BaggageItem(key string) string                         { return "" }
func (n testSpan) LogEvent(event string)                                 {}
func (n testSpan) LogEventWithPayload(event string, payload interface{}) {}
func (n testSpan) Log(data LogData)                                      {}

// StartSpan 实现 Tracer 接口
func (n testTracer) StartSpan(operationName string, opts ...StartSpanOption) Span {
	sso := StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return n.startSpanWithOptions(operationName, sso)
}

func (n testTracer) startSpanWithOptions(name string, opts StartSpanOptions) Span {
	fakeID := nextFakeID()
	if len(opts.References) > 0 {
		fakeID = opts.References[0].ReferencedContext.(testSpanContext).FakeID
	}

	return testSpan{
		OperationName: name,
		StartTime:     opts.StartTime,
		Tags:          opts.Tags,
		spanContext: testSpanContext{
			HasParent: len(opts.References) > 0,
			FakeID:    fakeID,
		},
	}
}

// Inject 实现 Tracer 接口
func (n testTracer) Inject(sp SpanContext, format interface{}, carrier interface{}) error {
	spanContext := sp.(testSpanContext)
	switch format {
	case HTTPHeaders, TextMap:
		carrier.(TextMapWriter).Set(testHTTPHeaderPrefix+"fakeid", strconv.Itoa(spanContext.FakeID))
		return nil
	}
	return ErrUnsupportedFormat
}

// Extract 实现 Tracer 接口
func (n testTracer) Extract(format interface{}, carrier interface{}) (SpanContext, error) {
	switch format {
	case HTTPHeaders, TextMap:
		// 目的仅仅是测试... 一般不值得真正去传播。
		sm := testSpanContext{}
		err := carrier.(TextMapReader).ForeachKey(func(key, val string) error {
			switch strings.ToLower(key) {
			case testHTTPHeaderPrefix + "fakeid":
				i, err := strconv.Atoi(val)
				if err != nil {
					return err
				}
				sm.FakeID = i
			}
			return nil
		})
		return sm, err
	}
	return nil, ErrSpanContextNotFound
}
