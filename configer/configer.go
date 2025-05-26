package configer

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// 读取配置

const ServiceName = "service.name"
const ServiceAddress = "service.address"
const RedisAddr = "redis.addr"
const RedisPassword = "redis.password"
const RedisDB = "redis.db"
const DBUrl = "db.url"
const KafkaBrokers = "kafka.brokers"
const KafkaLogMsgTopic = "kafka.msgTopic"
const kafkaErrMsgTopic = "kafka.errTopic"
const LoggerConsoleEnable = "logger.console.enable"
const LoggerFileEnable = "logger.file.enable"
const LoggerKafkaEnable = "logger.kafka.enable"
const LoggerFilePath = "logger.file.path"
const LoggerFileMaxAge = "logger.file.maxAge"
const LoggerFileMaxSize = "logger.file.maxSize"
const LoggerFileMaxBackups = "logger.file.maxBackups"
const LoggerFileCompress = "logger.file.compress"
const LoggerEnableLevel = "logger.enableLevel"
const ConsulAddress = "consul.address"
const Env = "env"

// ConsulViper viper with consul
var ConsulViper = viper.New()
var DefaultViper = viper.New()

type Config struct {
	// service
	ServiceName string // ServiceName `ServiceName`

	// service address
	ServiceAddress string // 服务地址，格式    :port

	// cache
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// db
	DBUrl string // 数据库连接地址

	// kafka
	KafkaBrokers     string
	KafkaLogMsgTopic string
	KafkaErrMsgTopic string

	// logger
	LoggerConsoleEnable bool
	LoggerFileEnable    bool
	LoggerKafkaEnable   bool
	LogFileMaxAge       int
	LogFileMaxSize      int
	LogFileMaxBackups   int
	LogFilePath         string
	LogFileCompress     bool
	LoggerEnableLevel   string
}

func Init() *Config {
	// parse commandline args
	env := flag.String("e", "development", "specified env. default is development")
	consulAddress := flag.String("consul_address", "", "specified consul address")
	flag.Parse()

	DefaultViper.Set("env", env)
	DefaultViper.Set(ConsulAddress, consulAddress)
	DefaultViper.AddConfigPath("configs")
	DefaultViper.SetConfigName(*env)
	err := DefaultViper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	consulConfigPath := "apps/" + DefaultViper.GetString(ServiceName) + "/configs"
	// init consul viper
	if err2 := ConsulViper.AddRemoteProvider("consul", DefaultViper.GetString(ConsulAddress), consulConfigPath); err2 != nil {
		//panic(fmt.Errorf("Fatal error consul: %w \n", err2))
		_ = fmt.Errorf("add remote provider failed: %w \n", err2)
	}
	ConsulViper.SetConfigType("json")
	readConsulErr2 := ConsulViper.ReadRemoteConfig()
	if readConsulErr2 != nil {
		//panic(readConsulErr2)
		_ = fmt.Errorf("read remote config failed: %w \n", readConsulErr2)
	}

	return &Config{
		ServiceName:    GetString(ServiceName),
		ServiceAddress: GetString(ServiceAddress),
		// redis
		RedisAddr:     GetString(RedisAddr),
		RedisPassword: GetString(RedisPassword),
		RedisDB:       GetInt(RedisDB),
		// db
		DBUrl: GetString(DBUrl),
		// kafka
		KafkaBrokers:     GetString(KafkaBrokers),
		KafkaLogMsgTopic: GetString(KafkaLogMsgTopic),
		KafkaErrMsgTopic: GetString(kafkaErrMsgTopic),
		// log
		LoggerConsoleEnable: GetBool(LoggerConsoleEnable),
		LoggerFileEnable:    GetBool(LoggerFileEnable),
		LoggerKafkaEnable:   GetBool(LoggerKafkaEnable),
		LogFileMaxAge:       GetInt(LoggerFileMaxAge),
		LogFileMaxSize:      GetInt(LoggerFileMaxSize),
		LogFileMaxBackups:   GetInt(LoggerFileMaxBackups),
		LogFilePath:         GetString(LoggerFilePath),
		LogFileCompress:     GetBool(LoggerFileCompress),
		LoggerEnableLevel:   GetString(LoggerEnableLevel),
	}
}

func GetString(key string) string {
	isSet := ConsulViper.IsSet(key)
	if isSet {
		return ConsulViper.GetString(key)
	}

	return DefaultViper.GetString(key)
}

func GetInt(key string) int {
	if ConsulViper.IsSet(key) {
		return ConsulViper.GetInt(key)
	}
	return DefaultViper.GetInt(key)
}

func GetBool(key string) bool {
	if ConsulViper.IsSet(key) {
		return ConsulViper.GetBool(key)
	}
	return DefaultViper.GetBool(key)
}
