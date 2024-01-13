package logger

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"time"
)

const (
	RFC3339Millis = "2006-01-02T15:04:05.000Z07:00"
)

func NewGinLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{Formatter: logFormatter})
}

func logFormatter(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	if param.ErrorMessage == "" {
		return fmt.Sprintf("{\"timestamp\":\"%v\", \"status_code\": \"%d\", \"latency\": \"%v\", \"latency_raw\": \"%d\", \"request_size\": \"%s\", \"request_size_raw\": \"%d\", \"client_ip\":\"%s\", \"method\": \"%s\", \"path\": \"%v\"}\n",
			param.TimeStamp.Format(RFC3339Millis),
			param.StatusCode,
			param.Latency,
			param.Latency,
			humanize.Bytes(uint64(param.BodySize)),
			param.BodySize,
			param.ClientIP,
			param.Method,
			param.Path,
		)
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
