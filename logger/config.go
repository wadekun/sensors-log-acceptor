package logger

import "go.uber.org/zap/zapcore"

// Config 日志配置
type Config struct {
	ServiceName         string        // 服务名称
	EnableLogLevel      zapcore.Level // 开启的日志级别
	ConsoleOutputEnable bool          // 开启输出到控制台
	FileOutputEnable    bool          // 开启输出到文件
	KafkaOutputEnable   bool          // 开启输出到kafka
	FileConfig          FileConfig    // 文件配置
	KafkaConfig         KafkaConfig   // kafka配置
}

// FileConfig 文件配置
type FileConfig struct {
	FilePath   string // 文件路径
	MaxSize    int    // 单个日志文件最大大小，单位 MB
	MaxBackups int    // 保留历史日志文件最大个数
	MaxAge     int    // 保留历史文件最大天数
	Compress   bool   // 是否压缩/归档历史日志文件
}

// KafkaConfig kafka配置
type KafkaConfig struct {
	brokers string // broker地址
	topic   string // topic
}
