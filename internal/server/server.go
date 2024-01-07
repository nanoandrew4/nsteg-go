package server

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"time"

	_ "nsteg/docs"
)

const (
	RFC3339Millis = "2006-01-02T15:04:05.000Z07:00"
)

// StartServer godoc
// @title nSteg API
// @version 1.0
// @description An API to perform steganography on images
// @BasePath /api/v1
func StartServer(port string) {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{Formatter: logFormatter}), gin.Recovery())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	v1.POST("/encode/image", EncodeImageHandler)

	http.HandleFunc("/encode/image", handleImageEncodeRequest)

	r.Run(fmt.Sprintf(":%s", port))
}

func logFormatter(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	// TODO: validate that JSON produced is valid with marshal/unmarshal test
	return fmt.Sprintf("{\"timestamp\":\"%v\", \"status_code\": \"%d\", \"latency\": \"%v\", \"latency_raw\": \"%d\", \"request_size\": \"%s\", \"request_size_raw\": \"%d\", \"client_ip\":\"%s\", \"method\": \"%s\", \"path\": \"%v\", \"error\": \"%s\"}\n",
		param.TimeStamp.Format(RFC3339Millis),
		param.StatusCode,
		param.Latency,
		param.Latency,
		humanize.Bytes(uint64(param.BodySize)),
		param.BodySize,
		param.ClientIP,
		param.Method,
		param.Path,
		param.ErrorMessage,
	)
}
