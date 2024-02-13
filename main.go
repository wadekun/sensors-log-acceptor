package main

import (
	"liangck.xyz/data-service/sensors-log-acceptor/cache"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"liangck.xyz/data-service/sensors-log-acceptor/dao"
	"liangck.xyz/data-service/sensors-log-acceptor/kafka"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
)

func main() {
	config := configer.Init()

	// init logger
	logger.Init(config)
	logger.Logger.Info("env: " + configer.GetString(configer.Env) + " , consulAddress: " + configer.GetString(configer.ConsulAddress))

	// init db resource
	dao.InitDb(config)
	logger.Logger.Info("init database resources successful.")

	// init cache resource
	cache.Init(config)
	logger.Logger.Info("init cache resources successful.")

	// init kafka
	kafka.Init(config)

	// init handler mapping and start gin
	InitRouter(config)
}
