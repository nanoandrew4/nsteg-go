package image

import (
	"image/png"
	"io"
	"nsteg/internal/bits"
	"nsteg/pkg/config"
	"testing"
)

func TestEncodeWithOpaqueImage(t *testing.T) {
	testEncode(t, false)
}

func TestEncodeWithPartiallyOpaqueImage(t *testing.T) {
	testEncode(t, true)
}

func testEncode(t *testing.T, randomizePixelOpaqueness bool) {
	img, opaquePixels := generateImage(ImageSize, ImageSize, randomizePixelOpaqueness)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		testFiles := generateFilesToEncode((opaquePixels * int(LSBsToUse) * 3) / 8)

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

		err = encoder.EncodeFiles(convertTestInputToStandardInput(testFiles), io.Discard)
		if err != nil {
			t.Fatalf("Error encoding files %s", err)
		}

		outputImage := encoder.image
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
			continue
		}

		testBitReader := bits.NewBitReader(expectedEncodedBytes)
		for currentPixel := firstOpaquePixelIdx + 1; testBitReader.BytesLeftToRead() > 0; currentPixel++ {
			px := currentPixel % outputImage.Bounds().Dx()
			py := currentPixel / outputImage.Bounds().Dx()
			pixelOffset := outputImage.PixOffset(px, py)
			pixel := outputImage.Pix[pixelOffset : pixelOffset+4]
			for channelIdx := byte(0); channelIdx < channelsToWrite; channelIdx++ {
				if pixel[3] == 255 {
					bitsToCheck := pixel[channelIdx] & (1<<LSBsToUse - 1)
					expectedBits := testBitReader.ReadBits(uint(LSBsToUse))
					if bitsToCheck != expectedBits {
						t.Errorf("Error with %d LSBs in pixel|channel %d|%d, expected|got %d|%d", LSBsToUse, currentPixel+1, channelIdx+1, expectedBits, bitsToCheck)
						testBitReader.Reset()
						break
					}
				}
			}
		}
	}
}
