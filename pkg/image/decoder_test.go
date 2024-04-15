package image

import (
	"bytes"
	"fmt"
	"nsteg/pkg/config"
	"nsteg/test"
	"testing"
)

func BenchmarkDecodeSpeed(b *testing.B) {
	for _, randomizePixelOpaqueness := range []bool{false, true} {
		b.Run(getOpaquenessLabel(randomizePixelOpaqueness), func(b *testing.B) {
			img, opaquePixels := generateImage(benchImageSize, benchImageSize, randomizePixelOpaqueness)
			for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
				numOfBytesToEncode := calculateBytesThatFitInImage(opaquePixels, LSBsToUse)
				bytesToEncode := test.GenerateRandomBytes(numOfBytesToEncode)
				iConfig := config.ImageEncodeConfig{
					LSBsToUse: LSBsToUse,
				}
				testImageEncoder, err := NewImageEncoder(img, iConfig)
				if err != nil {
					b.Fatalf("Error creating image encoder for benchmark")
				}
				err = testImageEncoder.Encode(bytes.NewReader(bytesToEncode))
				if err != nil {
					b.Fatalf("Error during image encoding: %s", err)
				}

				b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
					b.SetBytes(int64(numOfBytesToEncode))
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						testImageDecoder := Decoder{
							image:     img,
							LSBsToUse: LSBsToUse,
						}
						b.StartTimer()
						_, err = testImageDecoder.Decode(numOfBytesToEncode)
						if err != nil {
							b.Fatalf("Error during image decode: %s", err)
						}
					}
				})
			}
		})
	}
}
