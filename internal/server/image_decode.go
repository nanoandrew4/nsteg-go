package server

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"image"
	"io"
	"net/http"
	"nsteg/api"
	nstegImage "nsteg/pkg/image"
)

var (
	errDecode = api.Error{Code: "decode_error", Error: "error while decoding files from image"}
)

// DecodeImageHandler godoc
//
// @Summary Decode data from an image
// @Description This endpoint will decode the data previously encoded in the supplied image. The success response format is dictated by the Accept header, but all errors are returned as JSON
// @Tags image
// @Accept json,octet-stream
// @Produce json,octet-stream
// @Param requestBody body api.DecodeImageRequest true "Body with image to decode"
// @Success 200 {object} api.DecodeImageResponse
// @Failure 400 {object} api.Error
// @Failure 500 {object} api.Error
// @Router /decode/image [post]
func DecodeImageHandler(ctx *gin.Context) {
	var requestBody api.DecodeImageRequest

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errRequestBodyDecode)
		return
	}

	rawImageToDecode, _, err := image.Decode(io.NopCloser(bytes.NewReader(requestBody.ImageToDecode)))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errInvalidImage)
		return
	}

	rgbaImageToDecode, castOk := rawImageToDecode.(*image.RGBA)
	if !castOk {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errErrorConvertingRawImage)
		return
	}

	imageDecoder := nstegImage.NewImageDecoder(rgbaImageToDecode)
	decodedFiles, err := imageDecoder.DecodeFiles()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errDecode)
	}

	ctx.JSON(http.StatusOK, api.DecodeImageResponse{DecodedFiles: decodedFiles})
}
