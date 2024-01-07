package server

import (
	"bytes"
	"github.com/gin-gonic/gin"
	flatbuffers "github.com/google/flatbuffers/go"
	"image"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"nsteg/api"
	"nsteg/api/nsteg/EncodeImage"
	"nsteg/internal"
	"nsteg/pkg/encoder"
)

// EncodeImageHandler godoc
//
// @Summary Encode files into supplied image
// @Description This endpoint will encode the supplied files into the image, and return the encoded image. The success response format is dictated by the Accept header, but all errors are returned as JSON
// @Tags image
// @Accept json,octet-stream
// @Produce json,octet-stream
// @Param requestBody body api.EncodeImageRequest true "Body with image to encode and files to encode within the image, as well as configuration for the encoding process"
// @Success 200 {object} api.EncodeImageResponse
// @Failure 400 {object} api.Error
// @Failure 500 {object} api.Error
// @Router /encode/image [post]
func EncodeImageHandler(ctx *gin.Context) {
	var requestBody api.EncodeImageRequest

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{Error: "Error reading request body"})
		return
	}

	imageToEncode, _, err := image.Decode(bytes.NewReader(requestBody.ImageToEncode))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, api.Error{Code: "invalid_image", Error: "Supplied image to encode is invalid"})
		return
	}

	rgbaImg := image.NewRGBA(imageToEncode.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), imageToEncode, rgbaImg.Bounds().Min, draw.Src)
	imageToEncode = nil

	imageEncoder := encoder.NewImageEncoder(rgbaImg, internal.ImageEncodeConfig{
		LSBsToUse:           requestBody.LsbsToUse,
		PngCompressionLevel: png.BestCompression, // to reduce bandwidth costs since lower compression results in huge images
	})

	var filesToHide []internal.FileToHide
	for _, reqFileToHide := range requestBody.FilesToHide {
		filesToHide = append(filesToHide, internal.FileToHide{
			Name:    reqFileToHide.Name,
			Size:    int64(len(reqFileToHide.Content)),
			Content: bytes.NewReader(reqFileToHide.Content),
		})
	}

	encodedImageBuffer := bytes.NewBuffer(make([]byte, 0, len(requestBody.ImageToEncode))) // pre allocate with size of original, since it should be similar
	err = imageEncoder.EncodeFiles(filesToHide, encodedImageBuffer)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{Code: "encode_error", Error: "An error occurred while encoding the image"})
		return
	}

	ctx.JSON(http.StatusOK, api.EncodeImageResponse{EncodedImage: encodedImageBuffer.Bytes()})
}

func handleImageEncodeRequest(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading body", http.StatusInternalServerError)
		return
	}

	encodeImageRequest := EncodeImage.GetRootAsImageEncodeRequest(requestBody, 0)
	imageToEncodeSize := encodeImageRequest.ImageToEncodeLength()
	imageToEncode, _, err := image.Decode(bytes.NewReader(encodeImageRequest.ImageToEncodeBytes()))
	if err != nil {
		http.Error(w, "supplied image is invalid", http.StatusBadRequest)
		return
	}

	rgbaImg := image.NewRGBA(imageToEncode.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), imageToEncode, rgbaImg.Bounds().Min, draw.Src)
	imageToEncode = nil

	imageEncoder := encoder.NewImageEncoder(rgbaImg, internal.ImageEncodeConfig{
		LSBsToUse:           encodeImageRequest.LsbsToUse(),
		PngCompressionLevel: png.BestCompression, // to reduce bandwidth costs since lower compression results in huge images
	})

	var filesToHide []internal.FileToHide
	for i := 0; i < encodeImageRequest.FilesToHideLength(); i++ {
		var fbFileToHide EncodeImage.FileToHide
		read := encodeImageRequest.FilesToHide(&fbFileToHide, i)
		if !read {
			http.Error(w, "could not read file to hide", http.StatusInternalServerError)
			return
		}

		filesToHide = append(filesToHide, internal.FileToHide{
			Name:    string(fbFileToHide.Name()),
			Size:    int64(fbFileToHide.ContentLength()),
			Content: bytes.NewReader(fbFileToHide.ContentBytes()),
		})
	}
	encodeImageRequest = nil

	encodedImageBuffer := bytes.NewBuffer(make([]byte, 0, imageToEncodeSize)) // pre allocate with size of original, since it should be similar
	err = imageEncoder.EncodeFiles(filesToHide, encodedImageBuffer)
	if err != nil {
		http.Error(w, "error encoding image", http.StatusInternalServerError)
		return
	}

	fbResponseBuilder := flatbuffers.NewBuilder(imageToEncodeSize)

	EncodeImage.ImageEncodeResponseStart(fbResponseBuilder)
	offset := fbResponseBuilder.CreateByteVector(encodedImageBuffer.Bytes())
	EncodeImage.ImageEncodeResponseAddEncodedImage(fbResponseBuilder, offset)
	response := EncodeImage.ImageEncodeResponseEnd(fbResponseBuilder)
	fbResponseBuilder.Finish(response)
	_, err = w.Write(fbResponseBuilder.FinishedBytes())
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}
