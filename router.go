package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"liangck.xyz/data-service/sensors-log-acceptor/cache"
	"liangck.xyz/data-service/sensors-log-acceptor/configer"
	"liangck.xyz/data-service/sensors-log-acceptor/logger"
	"liangck.xyz/data-service/sensors-log-acceptor/middleware"
	"net/http"
)

// handle request
// 1.read request body
// 2.valid logger by meta data
// 3.send result to kafka
func handle(context *gin.Context) {
	jsonData, err := ioutil.ReadAll(context.Request.Body)
	logger.Logger.Info("origin jsonData: " + string(jsonData))
	if err != nil {
		context.JSON(http.StatusOK, gin.H{
			"errno": "1",
			"err":   err.Error(),
		})
	} else {
		ok, err2 := Handle(jsonData)
		if !ok && err2 != nil {
			context.JSON(http.StatusOK, gin.H{
				"errno": "1",
				"err":   err2.Error(),
			})
		} else {
			context.JSON(http.StatusOK, gin.H{
				"errno": "0",
			})
		}
	}
}

func InitRouter(config *configer.Config) {
	r := gin.Default()
	r.Use(middleware.GinLogger(logger.Logger), middleware.GinRecovery(logger.Logger, true))
	r.POST("/sa.go", handle)
	r.POST("/fieldChange", func(c *gin.Context) {
		cache.SendFieldChangeMessage()
		c.JSON(http.StatusOK, gin.H{
			"errno": "0",
		})
	})
	r.POST("/eventChange", func(c *gin.Context) {
		cache.SendEventChangeMessage()
		c.JSON(http.StatusOK, gin.H{
			"errno": "0",
		})
	})
	r.POST("/eventFieldChange", func(c *gin.Context) {
		event := c.Query("event")
		if event != "" {
			cache.SendEventFieldChangeMessage(event)
		}
		c.JSON(http.StatusOK, gin.H{
			"errno": "0",
		})
	})
	r.POST("/fieldValuesChange", func(c *gin.Context) {
		field := c.Query("field")
		if field != "" {
			cache.SendFieldValuesChangeMessage(field)
		}
		c.JSON(http.StatusOK, gin.H{
			"errno": "0",
		})
	})

	//r.GET("/getConfig", func(c *gin.Context) {
	//	key := c.Query("key")
	//	var value = ""
	//	if key != "" {
	//		value = configer.GetString(key)
	//	}
	//
	//	c.JSON(http.StatusOK, gin.H{
	//		"value": value,
	//	})
	//})

	// health check
	r.Any("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"error":    "",
			"errno":    "0",
			"dataType": "OBJECT",
			"data":     "",
		})
	})

	r.Run(config.ServiceAddress)
}
