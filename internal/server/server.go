package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	_ "nsteg/docs"
	"nsteg/internal/logger"
)

// StartServer godoc
// @title nSteg API
// @version 1.0
// @description An API to perform steganography on images
// @BasePath /api/v1
func StartServer(port string) {
	r := gin.New()
	r.Use(logger.NewGinLogger(), gin.Recovery())
	//pprof.Register(r, "debug/pprof")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	v1.POST("/image/encode", EncodeImageHandler)
	v1.POST("/image/decode", DecodeImageHandler)

	http.HandleFunc("/encode/image", handleImageEncodeRequest)

	r.Run(fmt.Sprintf(":%s", port))
}
