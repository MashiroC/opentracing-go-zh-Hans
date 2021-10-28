package opentracing

import (
	"time"

	"github.com/opentracing/opentracing-go/log"
)

// SpanContext 代表了 Span 的状态，它必须跨越进程边界，传播到下一世代的Span (child span)中。
// 例如，一个包含有 <trace_id, span_id,sampled(抽样，其他span包含的简单的信息)> 的元组。
type SpanContext interface {
	// ForeachBaggageItem 提供了了对于在 SpanContext 中的所有携带数据(baggage items)的访问方式。
	// handler 方法将对每一个键值对进行一次调用，但不保证调用顺序。
	//
	// bool 类型的返回值决定 handler 方法是否继续遍历剩下的携带数据；
	// 例如，如果handler想要根据某种规则查找一个携带数据中的键值对，
	// 它可以在找到数据的时候返回 false 来停止接下来的遍历
	ForeachBaggageItem(handler func(k, v string) bool)
}

// Span 接口的实例代表一个在 OpenTracing 系统中活跃的，未完成的 span
//
// Span 的实例由 Tracer 接口的实例创建。
type Span interface {
	// Finish 方法设置了结束的时间戳以及确定 Span 的最终状态。
	//
	// Finish() 必须是除了 Context() 之外对于 Span 的最后一个方法调用，
	// 否则将会导致未定义的行为。（Context() 可以在任何时候调用）
	Finish()
	// FinishWithOptions 就像是 Finish() 但是明确控制时间戳和日志数据。
	// （注：原文是：
	//      FinishWithOptions is like Finish() but with explicit control over
	//      timestamps and log data.
	//     我在看了Jaeger的实现之后，Jaeger的做法是把日志添加在后面，而不是覆盖或控制已有的日志。）
	FinishWithOptions(opts FinishOptions)

	// Context 返回该 Span 的 SpanContext。注意在 Span 调用 Span.Finish() 之后该方法仍然可用。
	Context() SpanContext

	// SetOperationName 可以设置或改变该 Span 的名字
	//
	// 返回一个该 Span 的指针
	SetOperationName(operationName string) Span

	// SetTag 会在当前 Span 添加一个 Tag。
	//
	// 如果该 Tag 的 key 已存在，将会覆盖旧的tag
	//
	// Tag 的值类型可以是数字，字符串或布尔值。其他类型的值在 OpenTracing 层面是未定义行为。
	// 如果一个特定的链路追踪系统不知道怎么处理一个特定类型的值，它可能会忽略该 Tag，但是不应该panic
	//
	// 返回一个该 Span 的指针
	SetTag(key string, value interface{}) Span

	// LogFields 提供了一个高效并且带有类型检查的方式来在 Span 下记录键值对，
	// 尽管接口相比 LogKV() 要啰嗦一点。这里是一个例子：
	//
	//    span.LogFields(
	//        log.String("event", "soft error"),
	//        log.String("type", "cache timeout"),
	//        log.Int("waited.millis", 1500))
	//
	// 也可以看看 Span.FinishWithOptions() 和 FinishOptions.BulkLogData.
	LogFields(fields ...log.Field)

	// LogKV 提供了一个简洁并可读的方式来在 Span 下记录键值对，
	// 不幸的是它相比 LogFiedls 的效率会低一些并且具有更弱的类型安全。这里是一个例子：
	//
	//    span.LogKV(
	//        "event", "soft error",
	//        "type", "cache timeout",
	//        "waited.millis", 1500)
	//
	// 对于 LogKV (对于 LogFields 相反)，参数必须是成对的键值对，就像下面这样：
	//
	//    span.LogKV(key1, val1, key2, val2, key3, val3, ...)
	//
	// 键必须是字符串，值可以是字符串、数字、布尔值、Go的错误或其他任意结构体。
	//
	// (对于该接口实现者的提示：考虑helper函数 log.InterleavedKVToFields() ）
	LogKV(alternatingKeyValues ...interface{})

	// SetBaggageItem 可以对于该 Span 设置一组键值对，
	// 并且它的 SpanContext 将传递该键值对到该 Span 的下一世代。
	//
	// SetBaggageItem() 提供了强大的全链路追踪集成，
	//（例如：来自移动段端的任意应用数据可以透明的一直进入到存储系统的底层。）
	// 但也有很高的成本，请在使用该功能的时候小心一些。
	//
	// 实现提示#1： SetBaggageItem() 仅将携带的数据传递给**下一时代**的 Span
	//
	// 实现提示#2: 使用该功能前请小心并且深思熟虑。
	// 每一个key和value将在你每一个本地**和远程**的子时代 Span中复制和传递，
	// 这个操作会大量增加你的网络和CPU占用。
	//
	// 返回一个该 Span 的指针
	SetBaggageItem(restrictedKey, value string) Span

	// BaggageItem 根据key获取携带数据的值。如果没有找到key对应的值，将会返回一个空字符串。
	BaggageItem(restrictedKey string) string

	// Tracer 提供了一个获取创建该 Span 的 Tracer 的方式。
	Tracer() Tracer

	// Deprecated: 弃用，请使用 LogFields 或者 LogKV
	LogEvent(event string)
	// Deprecated: 弃用，请使用 LogFields 或者 LogKV
	LogEventWithPayload(event string, payload interface{})
	// Deprecated: 弃用，请使用 LogFields 或者 LogKV
	Log(data LogData)
}

// LogRecord 是于单个 Span 日志关联的数据结构。每个 LogRecord 实例必须至少指定其中一个成员参数。
type LogRecord struct {
	Timestamp time.Time
	Fields    []log.Field
}

// FinishOptions 允许 Span.FinishWIthOptions 调用来覆盖结束的时间戳，
// 并提供了一个批量接口来添加日志。
type FinishOptions struct {
	// FinishTime 会覆盖 Span 的结束时间，如果 FinishTime.IsZero()，会隐式的调用 time.Now()
	//
	// FinishTime 必须大于 Span 的开始时间(StartTime)（对于 StartSpanOptions 也是一样）。
	FinishTime time.Time

	// LogRecords 允许调用者用一个切片(slice)来包含多条 LogFields() 调用。切片可能为nil
	//
	// 所有的 LogRecord.TimeStamp 值都不能是 IsZero()（即，必须显示的指明）。
	// 它们必须全部大于 Span 的开始时间且小于结束时间（如果 FinishTime.IsZero(),
	// 那么 FinishTime 为time.Now()）。
	// 否则 FinishWithOptions() 行为是未定义的。
	//
	// 如果指定，调用方会在调用 FinishWithOptions() 后移交 LogRecords 的所有权
	//
	// 如果指定，废弃(DEPRECATED)的 BulkLogData 必须是 nil 或者空。
	LogRecords []LogRecord

	// BulkLogData 是废弃(DEPRECATED)的。
	BulkLogData []LogData
}

// LogData 已弃用
type LogData struct {
	Timestamp time.Time
	Event     string
	Payload   interface{}
}

// ToLogRecord 转换一个废弃的 LogData 到一个没有废弃的 LogRecord
// ToLogRecord converts a deprecated LogData to a non-deprecated LogRecord
func (ld *LogData) ToLogRecord() LogRecord {
	var literalTimestamp time.Time
	if ld.Timestamp.IsZero() {
		literalTimestamp = time.Now()
	} else {
		literalTimestamp = ld.Timestamp
	}
	rval := LogRecord{
		Timestamp: literalTimestamp,
	}
	if ld.Payload == nil {
		rval.Fields = []log.Field{
			log.String("event", ld.Event),
		}
	} else {
		rval.Fields = []log.Field{
			log.String("event", ld.Event),
			log.Object("payload", ld.Payload),
		}
	}
	return rval
}
