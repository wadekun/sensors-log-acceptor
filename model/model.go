package model

// ErrType 错误类型
type ErrType int

const (
	None              ErrType = iota // 无异常   开始生成枚举值，从0开始
	TypeMisMatch                     // 类型不匹配
	ValueTooLong                     // 数值超长
	ValueCannotBeNull                // 未赋值（不可为空）
	ValueNotExist                    // 值不存在
	EventUndefined                   // Event 不存在(元数据中未定义)
	ParsedFailed                     // 解析失败
	InvalidFormat                    // 无效的数据格式
)

// Log 埋点日志
type Log struct {
	Gzip     string
	DataList string
	Data     string
	Crc      string
}

// ValidResult 验证结果
type ValidResult struct {
	// 是否验证通过
	OK bool
	// 错误信息
	Err string
	// 错误类型
	ErrType ErrType
}

// ReportError 上报异常
type ReportError struct {
	Err     string
	ErrType ErrType
	Data    string
	Time    int64  // 序列化后的时间
	ID      string // id UUID
}
