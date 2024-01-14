package image

import (
	"crypto/sha512"
	"fmt"
	"image/png"
	"io"
	"nsteg/pkg/config"
	"nsteg/pkg/model"
	"testing"
)

const TestFilePrefix = "testfile_"
const ImageSize = 3000

func TestEncodeDecodeOnOpaqueImage(t *testing.T) {
	testEncodeDecode(t, false)
}

func TestEncodeDecodeOnPartiallyOpaqueImage(t *testing.T) {
	testEncodeDecode(t, true)
}

func testEncodeDecode(t *testing.T, randomizePixelOpaqueness bool) {
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		imageToEncode, opaquePixels := generateImage(ImageSize, ImageSize, randomizePixelOpaqueness)

		testFiles := generateFilesToEncode((opaquePixels * int(LSBsToUse) * 3) / 8)
		originalHashes := calculateInputFileHashes(testFiles)
		encoder, err := NewImageEncoder(imageToEncode, config.ImageEncodeConfig{
			LSBsToUse:           LSBsToUse,
			PngCompressionLevel: png.NoCompression,
		})
		if err != nil {
			t.Errorf("Error creating image encoder")
			continue
		}

		err = encoder.EncodeFiles(convertTestInputToStandardInput(testFiles), io.Discard)
		if err != nil {
			t.Fatalf("Error encoding image: %s", err)
		}

		decoder, err := NewImageDecoder(encoder.image)
		if err != nil {
			t.Errorf("Error creating image decoder")
			continue
		}

		decodedFiles, err := decoder.DecodeFiles()
		if err != nil {
			t.Errorf("Error decoding image with %d LSBs: %s", LSBsToUse, err)
			continue
		}
		decodedHashes := calculateOutputFileHashes(decodedFiles)

		for i := 0; i < len(testFiles); i++ {
			if originalHashes[i] != decodedHashes[i] {
				t.Errorf("Hash for file %d is not the same after decoding | using %d LSBs", i, LSBsToUse)
			}
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
