package image

import (
	"crypto/sha512"
	"fmt"
	"image/png"
	"nsteg/pkg/config"
	"nsteg/pkg/model"
	"testing"
)

const TestFilePrefix = "testfile_"
const ImageSize = 3000

func TestEncodeDecodeFiles(t *testing.T) {
	runImageTestsWithAllLSBsAndOpaquenessSettings(t, encodeDecodeFiles)
}

func encodeDecodeFiles(t *testing.T, LSBsToUse byte, randomizePixelOpaqueness bool) {
	imageToEncode, opaquePixels := generateImage(ImageSize, ImageSize, randomizePixelOpaqueness)

	testFiles := generateFilesToEncode(calculateBytesThatFitInImage(opaquePixels, LSBsToUse))
	originalHashes := calculateInputFileHashes(testFiles)
	encoder, err := NewImageEncoder(imageToEncode, config.ImageEncodeConfig{
		LSBsToUse:           LSBsToUse,
		PngCompressionLevel: png.NoCompression,
	})
	if err != nil {
		t.Errorf("Error creating image encoder")
	}

	err = encoder.EncodeFiles(convertTestInputToStandardInput(testFiles))
	if err != nil {
		t.Fatalf("Error encoding image: %s", err)
	}

	decoder, err := NewImageDecoder(encoder.image)
	if err != nil {
		t.Errorf("Error creating image decoder")
	}

	decodedFiles, err := decoder.DecodeFiles()
	if err != nil {
		t.Errorf("Error decoding image with %d LSBs: %s", LSBsToUse, err)
	}
	decodedHashes := calculateOutputFileHashes(decodedFiles)

	for i := 0; i < len(testFiles); i++ {
		if originalHashes[i] != decodedHashes[i] {
			t.Errorf("Hash for file %d is not the same after decoding | using %d LSBs", i, LSBsToUse)
		}
	}
}

func calculateInputFileHashes(file []testInputFile) []string {
	hashes := make([]string, len(file), len(file))

	for idx, file := range file {
		h := sha512.New()
		h.Write([]byte(file.Name))
		h.Write(file.Content)
		hashes[idx] = fmt.Sprintf("%x", h.Sum(nil))
	}

	return hashes
}

func calculateOutputFileHashes(file []model.OutputFile) []string {
	hashes := make([]string, len(file), len(file))

	for idx, file := range file {
		h := sha512.New()
		h.Write([]byte(file.Name))
		h.Write(file.Content)
		hashes[idx] = fmt.Sprintf("%x", h.Sum(nil))
	}

	return hashes
}
