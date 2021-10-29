package opentracing

import (
	"errors"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// 核心传播接口(CORE PROPAGATION INTERFACES):
///////////////////////////////////////////////////////////////////////////////

var (
	// ErrUnsupportedFormat 发生在调用 Tracer.Inject() 或 Tracer.Extrace() 时 `format` 字段该 Tracer 没有实现的情况下。
	ErrUnsupportedFormat = errors.New("opentracing: Unknown or unsupported Inject/Extract format")

	// ErrSpanContextNotFound 发生在调用 Tracer.Extract() 时传输了有问题的 `carrier` 字段或该值内信息不足的情况下。
	ErrSpanContextNotFound = errors.New("opentracing: SpanContext not found in Extract carrier")

	// ErrInvalidSpanContext 发生在当 SpanContext 没有准备好时就调用 Tracer.Inject() 的情况下。
	// （例如，当它是另一个 Tracer 实现创建但由当前的 Tracer调用时）
	ErrInvalidSpanContext = errors.New("opentracing: SpanContext type incompatible with tracer")

	// ErrInvalidCarrier 发生在调用 Tracer.Inject() 或 Tracer.Extract() 时，Tracer 的实现期望获得的值与实际传输值不同的情况下。
	ErrInvalidCarrier = errors.New("opentracing: Invalid Inject/Extract carrier")

	// ErrSpanContextCorrupted 发生在调用 Tracer.Extract() 时，传入了非预期的 `carrier`。
	ErrSpanContextCorrupted = errors.New("opentracing: SpanContext data corrupted in Extract carrier")
)

///////////////////////////////////////////////////////////////////////////////
// 内置传播格式(BUILTIN PROPAGATION FORMATS):
///////////////////////////////////////////////////////////////////////////////

// BuiltinFormat 是一些内置的值，用于在 opentracing 包中调用 Tracer.Inject() 和 Tracer.Extract() 时指定序列化的格式
type BuiltinFormat byte

const (
	// Binary 代表 SpanContexts 的序列化格式是不透明的二进制数据
	//
	// 对于 Tracer.Inject()：载体(carrier)必须是`io.Writer`
	//
	// 对于 Tracer.Extract()：载体(carrier)必须是`io.Reader`
	Binary BuiltinFormat = iota

	// TextMap 代表 SpanContexts 是一个值的类型为string的键值对。
	//
	// 不像 HTTPHeaders，TextMap 的序列化格式不限制key和value的字符集。
	//
	// 对于 Tracer.Inject()：载体(carrier)必须是`TextMapWriter`
	//
	// 对于 Tracer.Extract(): 载体(carrier)必须是`TextMapReader`
	TextMap

	// HTTPHeaders 代表 SPanContext 是一个 HTTP header。
	//
	// 不像  TextMap，HTTPHeaders 格式需要键和值都是有效的HTTP Header。
	// （即：字符的大小写不稳定，键中不允许使用特殊字符，值应该是URL-escaped的，等）
	//
	// 对于 Tracer.Inject()：载体(carrier)必须是`TextMapWriter`
	//
	// 对于 Tracer.Extract(): 载体(carrier)必须是`TextMapReader`
	//
	// 见 HTTPHeadersCarrier 以获取遵循 http.Header 实例进行存储的 TextMapWriter 和 TextMapReader 的实现。

	// 例如，对于 Inject()：
	//
	//    carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
	//    err := span.Tracer().Inject(
	//        span.Context(), opentracing.HTTPHeaders, carrier)
	//
	// Extract():
	//
	//    carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
	//    clientContext, err := tracer.Extract(
	//        opentracing.HTTPHeaders, carrier)
	//
	HTTPHeaders
)

// TextMapWriter 是 Inject() 需要的载体 TextMap 的内置传播格式。调用者可以用它来编码一个 SpanContext 用于传播。编码类型是unicode字符串组成的map
type TextMapWriter interface {
	// Set 可以设置一个键值对到载体(carrier)。多个 Set() 调用但是使用相同的键是未定义行为。
	//
	// 注意：TextMapWriter 在后端存储时可能包含与 SpanContext 无关的数据。
	// 因此，调用 Inject() 和 Extract() 时， TextMapWriter 和 TextMapReader 接口的实现必须就前缀或其他约定达成一致，以区分它们自己的键值对。
	Set(key, val string)
}

// TextMapReader 是 Extract() 需要的载体 TextMap 内置的传播格式。调用者可以用它来解码一个用于传播的 SpanContext。编码类型是unicode字符串组成的map
type TextMapReader interface {
	// ForeachKey 可以通过重复调用`handler`函数来访问 TextMap 的内容。
	// 如果任何一次`handler`调用返回了一个非空(non-nil)的错误，ForeachKey 将会终止并且返回该错误。
	//
	// 注意：TextMapReader 在后端存储时可能包含与 SpanContext 无关的数据。
	// 因此，调用 Inject() 和 Extract() 时， TextMapWriter 和 TextMapReader 接口的实现必须就前缀或其他约定达成一致，以区分它们自己的键值对。
	//
	// "foreach" 回调模式在某些情况下减少了不必要的复制，并且允许在读取时保持锁。
	ForeachKey(handler func(key, val string) error) error
}

// TextMapCarrier 提供了对 TextMapWriter 和 TextMapReader 使用的常规的 map[string]string
type TextMapCarrier map[string]string

// ForeachKey 实现 TextMapReader 接口。
func (c TextMapCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, v := range c {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Set 实现 TextMapWriter 接口。
func (c TextMapCarrier) Set(key, val string) {
	c[key] = val
}

// HTTPHeadersCarrier 同时满足 TextMapWriter 和 TextMapReader 接口。
//
// 服务端用例:
//
//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
//     clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
//
// 客户端用例:
//
//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
//     err := tracer.Inject(
//         span.Context(),
//         opentracing.HTTPHeaders,
//         carrier)
//
type HTTPHeadersCarrier http.Header

// Set 实现 TextMapWriter 接口。
func (c HTTPHeadersCarrier) Set(key, val string) {
	h := http.Header(c)
	h.Set(key, val)
}

// ForeachKey 实现 TextMapReader 接口。
func (c HTTPHeadersCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
