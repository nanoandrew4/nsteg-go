package image

import (
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"nsteg/internal/bits"
	"nsteg/pkg/config"
	"testing"
)

func TestEncodeFilesWithOpaqueImage(t *testing.T) {
	runImageTestsWithAllLSBsAndOpaquenessSettings(t, encodeFiles)
}

func TestSinglePassEncode(t *testing.T) {
	runImageTestsWithAllLSBsAndOpaquenessSettings(t, testEncode(false))
}

func TestMultiPassEncode(t *testing.T) {
	runImageTestsWithAllLSBsAndOpaquenessSettings(t, testEncode(true))
}

func encodeFiles(t *testing.T, LSBsToUse byte, randomizePixelOpaqueness bool) {
	img, opaquePixels := generateImage(ImageSize, ImageSize, randomizePixelOpaqueness)
	testFiles := generateFilesToEncode(calculateBytesThatFitInImage(opaquePixels, LSBsToUse))

	var expectedEncodedBytes []byte
	expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(len(testFiles))...)
	for _, file := range testFiles {
		expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(len(file.Name))...)
		expectedEncodedBytes = append(expectedEncodedBytes, []byte(file.Name)...)

		expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(len(file.Content))...)
		fileContent := file.Content
		expectedEncodedBytes = append(expectedEncodedBytes, fileContent...)
	}

	encoder, err := NewImageEncoder(img, config.ImageEncodeConfig{
		LSBsToUse:           LSBsToUse,
		PngCompressionLevel: png.NoCompression,
	})
	if err != nil {
		t.Fatalf("Error creating image encoder")
	}

	err = encoder.EncodeFiles(convertTestInputToStandardInput(testFiles))
	if err != nil {
		t.Fatalf("Error encoding files %s", err)
	}

	checkEncodedImageAgainstExpectedBytes(t, encoder.image, LSBsToUse, expectedEncodedBytes)
}

func testEncode(multiPass bool) testFunc {
	return func(t *testing.T, LSBsToUse byte, randomizePixelOpaqueness bool) {
		img, opaquePixels := generateImage(ImageSize, ImageSize, randomizePixelOpaqueness)
		encoder, err := NewImageEncoder(img, config.ImageEncodeConfig{LSBsToUse: LSBsToUse})
		if err != nil {
			t.Fatalf("Error creating image encoder")
		}

		var fullBytesToEncode []byte
		var encodesToPerform = 1
		if multiPass {
			encodesToPerform = rand.Intn(99) + 1
		}
		for i := 0; i < encodesToPerform; i++ {
			bytesToEncode := generateRandomBytes(calculateBytesThatFitInImage(opaquePixels, LSBsToUse) / encodesToPerform)
			err = encoder.Encode(bytes.NewReader(bytesToEncode))
			if err != nil {
				t.Fatalf("Error encoding files %s", err)
			}
			fullBytesToEncode = append(fullBytesToEncode, bytesToEncode...)
		}

		checkEncodedImageAgainstExpectedBytes(t, encoder.image, LSBsToUse, fullBytesToEncode)
	}
}

func checkEncodedImageAgainstExpectedBytes(t *testing.T, outputImage *image.RGBA, LSBsToUse byte,
	expectedEncodedBytes []byte) {

	var firstOpaquePixelIdx int
	var firstOpaquePixel [3]byte
	for p := 0; p < len(outputImage.Pix)/4; p++ {
		if outputImage.Pix[p*4+3] == 255 {
			firstOpaquePixelIdx = p
			firstOpaquePixel = [3]byte(outputImage.Pix[p*4 : p*4+3])
			break
		}
	}
	encodedLSBsToUse := (firstOpaquePixel[0] & 1) + (firstOpaquePixel[1]&1)<<1 + (firstOpaquePixel[2]&1)<<2 + 1
	if LSBsToUse != encodedLSBsToUse {
		t.Errorf("Expected encoded LSBsToUse to be %d, was %d", LSBsToUse, encodedLSBsToUse)
	}

	testBitReader := bits.NewBitReader(expectedEncodedBytes)
	for currentPixel := firstOpaquePixelIdx + 1; testBitReader.BytesLeftToRead() > 0; currentPixel++ {
		px := currentPixel % outputImage.Bounds().Dx()
		py := currentPixel / outputImage.Bounds().Dx()
		pixelOffset := outputImage.PixOffset(px, py)
		pixel := outputImage.Pix[pixelOffset : pixelOffset+4]
		for channelIdx := byte(0); channelIdx < channelsToWrite && testBitReader.BytesLeftToRead() > 0; channelIdx++ {
			var bitsToRead uint
			if testBitReader.BitsLeftToRead() >= int(LSBsToUse) {
				bitsToRead = uint(LSBsToUse)
			} else {
				bitsToRead = uint(testBitReader.BitsLeftToRead())
			}
			if pixel[3] == 255 {
				bitsToCheck := pixel[channelIdx] & (1<<bitsToRead - 1)
				expectedBits := testBitReader.ReadBits(bitsToRead)
				if bitsToCheck != expectedBits {
					t.Errorf("Error with %d LSBs when reading %d bits in pixel|channel %d|%d, expected|got %d|%d",
						LSBsToUse, bitsToRead, currentPixel+1, channelIdx+1, expectedBits, bitsToCheck)
					testBitReader.Reset()
					break
				}
			}
		}
	}
}
