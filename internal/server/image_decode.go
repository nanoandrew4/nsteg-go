package server

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"image"
	"net/http"
	"nsteg/api"
	"nsteg/internal/logging"
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

	logger := logging.BuildLoggerFromCtx(ctx)
	logger.Debug("Processing image decode request")

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		logger.WithError(err).Error("Error decoding request body")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errRequestBodyDecode)
		return
	}

	rawImageToDecode, _, err := image.Decode(bytes.NewReader(requestBody.ImageToDecode))
	if err != nil {
		logger.WithError(err).Error("Error decoding request image")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errInvalidImage)
		return
	}

	rgbaImageToDecode, castOk := rawImageToDecode.(*image.RGBA)
	if !castOk {
		logger.Error("Supplied image was not RGBA, casting failed")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errErrorConvertingRawImage)
		return
	}

	imageDecoder, err := nstegImage.NewImageDecoder(rgbaImageToDecode)
	if err != nil {
		handleDecodeError(ctx, logger, err)
		return
	}

	decodedFiles, err := imageDecoder.DecodeFiles()
	if err != nil {
		handleDecodeError(ctx, logger, err)
		return
	}

	logger.With("stats", toHumanizedDecodeStats(imageDecoder.Stats())).Info("Image decoding was successful")

	ctx.JSON(http.StatusOK, api.DecodeImageResponse{DecodedFiles: decodedFiles})
}

func handleDecodeError(ctx *gin.Context, logger *logging.Logger, err error) {
	logger.WithError(err).Error("Error decoding data from image")
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, errDecode)
}
