package kafka

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
	. "liangck.xyz/data-service/sensors-log-acceptor/model"
	"strings"
	"time"
)

var kafkaConf *Conf
var producer *Producer

// Init 初始化kafka config producer
func Init(config *configer.Config) {
	kafkaConf = &Conf{
		Brokers: config.KafkaBrokers,     // "192.168.3.212:9092",
		Topic:   config.KafkaLogMsgTopic, // "user_event_log",
	}
	producer = NewProducer(kafkaConf)
}

// Conf kafka configuration
type Conf struct {
	Topic   string `toml:"kafka_topic"`
	Brokers string `toml:"kafka_broker"`
}

type Producer struct {
	kafkaConf   *Conf
	kafkaWriter *kafka.Writer
}

// NewProducer create new kafka write instance
func NewProducer(kafkaConf *Conf) *Producer {
	producer := &Producer{
		kafkaConf: kafkaConf,
	}

	brokerArr := strings.Split(kafkaConf.Brokers, ",")

	w := &kafka.Writer{
		Addr: kafka.TCP(brokerArr...),
		//Topic:    kafkaConf.Topic,
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	producer.kafkaWriter = w
	return producer
}

// WriteErrorMsg 发送异常信息至异常信息Topic
func WriteErrorMsg(error *ReportError) {
	error.Time = time.Now().UnixMilli()
	newUUID, err3 := uuid.NewUUID()
	if err3 != nil {
		logger.Logger.Error("failed to get uuid : " + err3.Error())
	} else {
		error.ID = newUUID.String()
	}
	errorJson, err := json.Marshal(error)
	if err != nil {
		logger.Logger.Error("Failed to Marshal error msg . caused by: " + err.Error())
	}

	err2 := producer.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Topic: "user_event_log_err",
			Value: errorJson,
		},
	)

	if err2 != nil {
		logger.Logger.Error("Failed to send error msg to kafka. caused by: " + err2.Error())
	}
}

// WriteLogMsg 发送（验证通过的）上报行为日志数据
func WriteLogMsg(logMap *map[string]interface{}) {
	logJson, err := json.Marshal(logMap)
	if err != nil {
		logger.Logger.Error("Failed to Marshal log msg. caused by: " + err.Error())
	}

	err2 := producer.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Topic: kafkaConf.Topic,
			Value: logJson,
		},
	)

	if err2 != nil {
		logger.Logger.Error("Failed to send log msg to kafka. caused by: " + err2.Error())
	}
}
