package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"nsteg/pkg/model"
	"strconv"
	"testing"
)

type testInputFile struct {
	Name    string
	Content []byte
}

type testFunc func(t *testing.T, LSBsToUse byte, randomizePixelOpaqueness bool)

func runImageTestsWithAllLSBsAndOpaquenessSettings(t *testing.T, testFunc testFunc) {
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		LSBsToUseCopy := LSBsToUse
		t.Run(fmt.Sprintf("LSBsToUse-%d", LSBsToUse), func(t *testing.T) {
			t.Parallel()
			t.Run("opaque", func(t *testing.T) {
				t.Parallel()
				testFunc(t, LSBsToUseCopy, true)
			})
			t.Run("non-opaque", func(t *testing.T) {
				t.Parallel()
				testFunc(t, LSBsToUseCopy, false)
			})
		})
	}
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

func calculateBytesThatFitInImage(opaquePixels int, LSBsToUse byte) int {
	return ((opaquePixels - 1) * int(LSBsToUse) * 3) / 8
}

func generateFilesToEncode(availableBytes int) (testFiles []testInputFile) {
	var filesToEncode []testInputFile

	var exit bool
	var numOfBytesGenerated = 64 // 64 bits are needed to encode the number of files
	for i := 0; !exit; i++ {
		fileName := TestFilePrefix + strconv.Itoa(i)
		bytesToUseForFile := rand.Intn(availableBytes - (8 + len(fileName) + 8))

		// a file requires 8 bytes for the length of the name, plus however many bytes long the name is, plus eight bytes
		// for the file size, plus however many bytes the file is made up of
		bytesRequiredForNextFile := 8 + len(fileName) + 8 + bytesToUseForFile
		if numOfBytesGenerated+bytesRequiredForNextFile > availableBytes {
			bytesToUseForFile = availableBytes - numOfBytesGenerated - (8 + len(fileName) + 8)
			bytesRequiredForNextFile = 8 + len(fileName) + 8 + bytesToUseForFile
			exit = true
		}

		filesToEncode = append(filesToEncode, testInputFile{
			Name:    fileName,
			Content: generateRandomBytes(bytesToUseForFile),
		})

		numOfBytesGenerated += bytesRequiredForNextFile
	}

	return filesToEncode
}

func generateRandomBytes(numOfBytesToGenerate int) []byte {
	generatedBytes := make([]byte, numOfBytesToGenerate)
	_, err := rand.Read(generatedBytes)
	if err != nil {
		panic(err)
	}
	return generatedBytes
}

func getOpaquenessLabel(randomizeOpaqueness bool) string {
	if randomizeOpaqueness {
		return "non-opaque"
	} else {
		return "opaque"
	}
}
