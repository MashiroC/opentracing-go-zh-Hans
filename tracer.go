package opentracing

import "time"

// Tracer 是一个简单，轻量的用于创建 Span 和传递 SpanContext 的接口。
type Tracer interface {

	// StartSpan 创建并开始一个新的 Span，并且拥有指定的 `operationName（操作名）` 和 `StartSpanOption（启动选项）`.
	// （注意：参数`opt`使用了"函数选项"模式，可以通过这个链接查看该模式的详解）
	// （http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis）
	//
	// 如果一个Span在没有任何关联Span(SpanReference)的情况下（例如没有调用`opentracing.ChildOf()`或者`opentracing.FollowsFrom()`）创建，
	// 该Span将成为一个根Span（root span）
	//
	// 例子:
	//
	//     var tracer opentracing.Tracer = ...
	//
	//     // 根Span用例：
	//     sp := tracer.StartSpan("GetFeed")
	//
	//     (原文：The vanilla child span，经过咨询后get了vanilla还有nothing exciting的意思)
	//     // 一个平凡无奇的 child Span 用例：
	//     sp := tracer.StartSpan(
	//         "GetFeed",
	//         opentracing.ChildOf(parentSpan.Context()))
	//
	//     （原文：All the bells and whistles）
	//     // 点缀上更多的选项：
	//     sp := tracer.StartSpan(
	//         "GetFeed",
	//         opentracing.ChildOf(parentSpan.Context()),
	//         opentracing.Tag{"user_agent", loggedReq.UserAgent},
	//         opentracing.StartTime(loggedReq.Timestamp),
	//     )
	//
	StartSpan(operationName string, opts ...StartSpanOption) Span

	// Inject 将 SpanContext 注入到载体(carrier)中，载体的实际类型由格式(format)的值决定
	//
	// OpenTracing 定义了一组通用的`format`值（详见`BuildinFormat`），每个值都有一个预期的载体类型。
	//
	// 其他的包可能会定义它们自己的`format`值，就像是`context.Context`中的键一样，
	// （详见 https://godoc.org/context#WithValue）
	//
	// 例子（默认不处理错误）:
	//
	//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
	//     err := tracer.Inject(
	//         span.Context(),
	//         opentracing.HTTPHeaders,
	//         carrier)
	//
	// 注意：所有基于opentracing约定的Tracer实现都必须支持所有的`BuiltinFormats`
	//
	// 基于不同的实现，该函数在传入不支持的`format`值或未知的`format`值时，
	// 可能会返回错误 opentracing.ErrUnsupportedFormat
	//
	// 在`fotmat`的值是支持的但是注入失败，或其他基于特定实现的问题下，
	// 该函数可能会返回错误 opentracing.ErrInvalidCarrier
	//
	// 你可以再看看 Extract() (see Extract())
	Inject(sm SpanContext, format interface{}, carrier interface{}) error

	// Extract 通过传入的格式(format)和载体(carrier)，来返回一个SpanContext
	//
	// OpenTracing 定义了一组通用的`format`值（详见`BuildinFormat`），每个值都有一个预期的载体类型。
	//
	// 其他的包可能会定义它们自己的`format`值，就像是`context.Context`中的键一样，
	// （详见 https://godoc.org/context#WithValue）
	//
	// 例子（带有StartSpan）：
	//
	//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
	//     clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
	//
	//     // ... 假设这里的目标是继续客户端的追踪并且创建一个服务端的Span：
	//     var serverSpan opentracing.Span
	//     if err == nil {
	//         span = tracer.StartSpan(
	//             rpcMethodName, ext.RPCServerOption(clientContext))
	//     } else {
	//         span = tracer.StartSpan(rpcMethodName)
	//     }
	//
	// 注意：所有基于opentracing约定的Tracer实现都必须支持所有的`BuiltinFormats`
	//
	// 函数执行的返回值情况:
	//  - 一次成功的 Extract 调用会返回一个 SpanContext和nil
	//  - 如果在`carrier`中没有需要被提取的SpanContext，
	//    那么该函数会返回 (nil, opentracing.ErrSpanContextNotFound)
	//  - 如果`format`是不支持的或者无法识别，那么该函数会返回 (nil, opentracing.ErrUnsupportedFormat)
	//  - 如果`carrier`对象有一些更基本的问题(more fundamental problems，我觉得翻译成"其他的问题"会更好一点)，
	//    该函数可能会返回这些错误：opentracing.ErrInvalidCarrier, opentracing.ErrSpanContextCorrupted,
	//    或者其他的基于实现的特有的错误
	//
	// 你可以再看看 Inject() (see Inject())
	Extract(format interface{}, carrier interface{}) (SpanContext, error)
}

// StartSpanOptions 允许 Tracer.StartSpan 通过在调用中传递本结构体来实现某种机制，
// 比如覆盖Span开始时间的时间戳，指定一个Span的关联(Span References)，以及使一个或
// 多个Tag在Span开始时可用
//
// StartSpan() 调用应该查看在这个包中的`StartSpanOption`的接口和实现
//
// Tracer 的实现能将一个`StartSpanOption`的切片转化为一个`StartSpanOptions`的结构体，
// 就像下面这样：
//
//     func StartSpan(opName string, opts ...opentracing.StartSpanOption) {
//         sso := opentracing.StartSpanOptions{}
//         for _, o := range opts {
//             o.Apply(&sso)
//         }
//         ...
//     }
//
type StartSpanOptions struct {
	// Refenerces 存储了与其他Span相关联的SpanContext，长度可能为零。
	// 如果为空，创建一个新的 root span。（即开启一条新链路）
	References []SpanReference

	// StartTime 存储了该Span的开始时间，如果 StartTime.IsZero()，则该值默认为 time.Now()
	StartTime time.Time

	// Tags 可能包含多个值；对该 map 的值的限制与 Span.SetTag() 相同。该字段可能会为nil
	//
	// 在StartSpan调用之后请不要在其他地方使用该值
	Tags map[string]interface{}
}

// StartSpanOption 接口的实例可能会传给 Tracer.StartSpan.
//
// StartSpanOption 遵循了"函数选项(functional options)"模式，可以在这里了解该模式：
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type StartSpanOption interface {
	Apply(*StartSpanOptions)
}

// SpanReferenceType 是一个枚举类型，用于描述在两个有关联的Span中它们的关联类型。
// 如果 Span-2 和 Span1 有关系，那么 SpanReferenceType 会在 Span-2的角度来描述自己和 Span-1 的关系。
// 例如，ChildofRef 意味着 Span-1 创建了 Span-2，Span-1 是 Span-2 的上游。
//
// 注意：Span-1 和 Span-2 **并不一定** 是互相有依赖的；即，Span-2 可能是 Span-1 的后台任务之一，
// 或者 Span-2 可能是一个分布式队列中比 Span-1 稍后等待的任务。
type SpanReferenceType int

const (
	// ChildOfRef 描述了父Span(parent Span)和依赖于它的子Span(child Span)的关系。
	// 通常（但不是一定），在子Span完成前，父Span不能完成。
	//
	// 这个时序图描述了子Span阻塞了父Span的结束。
	//
	//     [-Parent Span---------]
	//          [-Child Span----]
	//
	// 详见 http://opentracing.io/spec/
	//
	// 你可以看看 opentracing.ChildOf()
	ChildOfRef SpanReferenceType = iota

	// FollowsFromRef 描述了一个父Span并不依赖于子Span执行结果的Span之间的关系。
	// 一般通常用`FollowsFromRefs`来描述一个由队列分割的流水线阶段，
	// 或一个在web请求的尾端插入的即发即忘缓存。
	//
	// 一个 FollowsFromRef Span 与新Span具有相同的逻辑，是同一个链路的一部分。
	// 即，新的Span不知道是由何原因被旧的Span创建的。
	//
	// 以下所有时序图都是FollowFromRef描述的关系的可能情况。
	//
	//     [-Parent Span-]  [-Child Span-]
	//
	//
	//     [-Parent Span--]
	//      [-Child Span-]
	//
	//
	//     [-Parent Span-]
	//                 [-Child Span-]
	//
	// 详见 http://opentracing.io/spec/
	//
	// 你可以看看 opentracing.FollowsFrom()
	FollowsFromRef
)

// SpanReference 是一个 StartSpanOption，包含了另一个 SpanContext 及该 Span 与另一个 Span 的关系。
// 在 SpanReferenceType 的文档详见支持的关系类型。如果 SpanReference 的值为空，则该结构体不起作用。
// 因此它允许使用一个更简单的方法来开始一个新的Span：
//
//     sc, _ := tracer.Extract(someFormat, someCarrier)
//     span := tracer.StartSpan("operation", opentracing.ChildOf(sc))
//
// 如果sc为空(sc == nil)，选项`ChildOf(sc)`不会panic，并且不会将父Span添加到options中。
type SpanReference struct {
	Type              SpanReferenceType
	ReferencedContext SpanContext
}

// Apply 实现 StartSpanOption 接口
func (r SpanReference) Apply(o *StartSpanOptions) {
	if r.ReferencedContext != nil {
		o.References = append(o.References, r)
	}
}

// ChildOf 返回一个指向依赖的父节点的SpanContext，已实现接口`StartSpanOption`。
// 如果sc为空(sc == nil)，该选项不起作用。
//
// 可以看看 ChildOfRef, SpanReference
func ChildOf(sc SpanContext) SpanReference {
	return SpanReference{
		Type:              ChildOfRef,
		ReferencedContext: sc,
	}
}

// FollowsFrom 返回一个指向父节点的SpanContext，已实现接口`StartSpanOption`。
// 父Span任何情况下都并不直接依赖于子Span的结果。
// 如果sc为空(sc == nil)，该选项不起作用
//
// 可以看看 FollowsFromRef, SpanReference
func FollowsFrom(sc SpanContext) SpanReference {
	return SpanReference{
		Type:              FollowsFromRef,
		ReferencedContext: sc,
	}
}

// StartTime 实现了`StartSpanOption`接口，用于对Span设置一个明确的开始时间
type StartTime time.Time

// Apply 实现`StartSpanOption`接口.
func (t StartTime) Apply(o *StartSpanOptions) {
	o.StartTime = time.Time(t)
}

// Tags 是一个通用的map，是string到不透明类型值(interface)的映射，
// 底层的链路追踪系统负责解释和序列化该Tag
type Tags map[string]interface{}

// Apply 实现`StartSpanOption`接口.
func (t Tags) Apply(o *StartSpanOptions) {
	if o.Tags == nil {
		o.Tags = make(map[string]interface{})
	}
	for k, v := range t {
		o.Tags[k] = v
	}
}

// Tag 可能会作为`StartSpanOption`来传递，为新的Span添加一个tag，
// 或者它的`Set`方法可能用于在一个已有的Span上添加新的tag
// 例如:
//
// tracer.StartSpan("opName", Tag{"Key", value})
//
//   或者
//
// Tag{"key", value}.Set(span)
type Tag struct {
	Key   string
	Value interface{}
}

// Apply 实现`StartSpanOption`接口.
func (t Tag) Apply(o *StartSpanOptions) {
	if o.Tags == nil {
		o.Tags = make(map[string]interface{})
	}
	o.Tags[t.Key] = t.Value
}

// Set 会在一个已有的Span上添加新的tag
func (t Tag) Set(s Span) {
	s.SetTag(t.Key, t.Value)
}
