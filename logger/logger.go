package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"os"
	"time"
)

// zap + gin: https://www.cnblogs.com/you-men/p/14694928.html#_labelTop

var Logger *zap.Logger
var SugarLogger *zap.SugaredLogger

// Init 初始化日志配置
func Init(config *configer.Config) {
	fileConfig := FileConfig{
		Compress:   config.LogFileCompress,
		MaxSize:    config.LogFileMaxSize,
		MaxBackups: config.LogFileMaxBackups,
		MaxAge:     config.LogFileMaxAge,
		FilePath:   config.LogFilePath,
	}
	logConf := Config{
		FileConfig:          fileConfig,
		EnableLogLevel:      zapcore.DebugLevel,
		ConsoleOutputEnable: config.LoggerConsoleEnable,
		FileOutputEnable:    config.LoggerFileEnable,
		KafkaOutputEnable:   config.LoggerKafkaEnable,
	}

	var allCore []zapcore.Core
	encoder := getEncoder(&logConf)
	if logConf.FileOutputEnable {
		writeSyncer := getFileWriteSyncer(&logConf)
		fileCore := zapcore.NewCore(encoder, writeSyncer, logConf.EnableLogLevel)
		allCore = append(allCore, fileCore)
	}
	if logConf.ConsoleOutputEnable {
		consoleWriter := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleWriter, logConf.EnableLogLevel)
		allCore = append(allCore, consoleCore)
	}
	// todo
	if logConf.KafkaOutputEnable {

	}
	core := zapcore.NewTee(allCore...)
	Logger = zap.New(core, zap.AddCaller())
	SugarLogger = Logger.Sugar()
}

func getFileWriteSyncer(config *Config) zapcore.WriteSyncer {
	logger := &lumberjack.Logger{
		Filename:   config.FileConfig.FilePath,
		MaxSize:    config.FileConfig.MaxSize,
		MaxBackups: config.FileConfig.MaxBackups,
		MaxAge:     config.FileConfig.MaxAge,
		Compress:   config.FileConfig.Compress,
	}

	return zapcore.AddSync(logger)
}

func getEncoder(config *Config) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
