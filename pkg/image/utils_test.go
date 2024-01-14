package image

import (
	"bytes"
	"image"
	"image/color"
	"math/rand"
	"nsteg/pkg/model"
	"strconv"
)

type testInputFile struct {
	Name    string
	Content []byte
}

func convertTestInputToStandardInput(testInputFiles []testInputFile) []model.InputFile {
	var inputFiles []model.InputFile
	for _, tif := range testInputFiles {
		inputFiles = append(inputFiles, model.InputFile{
			Name:    tif.Name,
			Content: bytes.NewReader(tif.Content),
			Size:    int64(len(tif.Content)),
		})
	}
	return inputFiles
}

func generateImage(width, height int, randomizePixelOpaqueness bool) (img *image.RGBA, opaquePixels int) {
	img = image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if randomizePixelOpaqueness && rand.Int()%(rand.Int()%4+1) == 0 {
				img.Set(x, y, color.RGBA{R: randUint8(), G: randUint8(), B: randUint8(), A: randUint8()})
			} else {
				opaquePixels++
				img.Set(x, y, color.RGBA{R: randUint8(), G: randUint8(), B: randUint8(), A: 255})
			}
		}
	}
	return img, opaquePixels
}

func randUint8() uint8 {
	return uint8(rand.Intn(256))
}

func generateFilesToEncode(availableBytes int) []testInputFile {
	var filesToEncode []testInputFile

	var numOfBytesGenerated int
	for i := 0; ; i++ {
		fileName := TestFilePrefix + strconv.Itoa(i)
		bytesToUseForFile := rand.Intn(availableBytes - (8 + len(fileName) + 8))

		// a file requires 8 bytes for the length of the name, plus however many bytes long the name is, plus eight bytes
		// for the file size, plus however many bytes the file is made up of
		bytesRequiredForNextFile := 8 + len(fileName) + 8 + bytesToUseForFile
		if numOfBytesGenerated+bytesRequiredForNextFile > availableBytes {
			return filesToEncode
		}

		generatedBytes := make([]byte, bytesToUseForFile)
		_, err := rand.Read(generatedBytes)
		if err != nil {
			panic(err)
		}

		filesToEncode = append(filesToEncode, testInputFile{
			Name:    fileName,
			Content: generatedBytes,
		})

		numOfBytesGenerated += bytesRequiredForNextFile
	}

	return filesToEncode
}
