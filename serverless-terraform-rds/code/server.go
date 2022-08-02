package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"serverless-terraform-rds/code/api"
	"serverless-terraform-rds/code/tool"
)

func Init() {
	viper.SetDefault("log-level", logrus.InfoLevel)
	_ = viper.BindEnv("log-level", "SERVERLESS_TERRAFORM_LOG_LEVEL")
}

var logger *logrus.Entry

func start() {
	var stopChan chan int
	r := gin.Default()
	r.POST("/invoke", func(ctx *gin.Context) {
		logger = tool.GetLoggerByRequestID(ctx.GetHeader("x-fc-request-id"))
		stop := make(chan int)
		// save stop for pre-freeze to use.
		stopChan = stop
		api.Invoke(ctx, logger, stop)
		stopChan = nil
	})

	r.GET("/pre-freeze", func(context *gin.Context) {
		api.PreFreeze(logger, stopChan, context)
	})
	if err := r.Run(":9000"); err != nil {
		panic(err)
	}
}
func main() {
	Init()
	start()
}
