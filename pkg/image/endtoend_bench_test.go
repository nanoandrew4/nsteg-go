package image

import (
	"fmt"
	"image/png"
	"io"
	"nsteg/pkg/config"
	"testing"
)

func BenchmarkFullEncodeSpeed(b *testing.B) {
	img, _ := generateImage(ImageSize, ImageSize, false)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, numOfBytesToEncode := range []int{100000, 1000000, 10000000} {
			filesToHide := convertTestInputToStandardInput(generateFilesToEncode(numOfBytesToEncode))
			encoder := NewImageEncoder(img, config.ImageEncodeConfig{
				LSBsToUse:           LSBsToUse,
				PngCompressionLevel: png.NoCompression,
			})
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(numOfBytesToEncode)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				for i := 0; i < b.N; i++ {
					_ = encoder.EncodeFiles(filesToHide, io.Discard)
				}
			})
		}
	}
}

func BenchmarkFullDecodeSpeed(b *testing.B) {
	img, _ := generateImage(ImageSize, ImageSize, false)

	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, numOfBytesToEncode := range []int{100000, 1000000, 10000000} {
			filesToHide := convertTestInputToStandardInput(generateFilesToEncode(numOfBytesToEncode))
			encoder := NewImageEncoder(img, config.ImageEncodeConfig{
				LSBsToUse:           LSBsToUse,
				PngCompressionLevel: png.NoCompression,
			})
			err := encoder.EncodeFiles(filesToHide, io.Discard)
			if err != nil {
				b.Fatalf("Error encoding file for decode benchmark: %s", err)
			}
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(numOfBytesToEncode)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					decoder := NewImageDecoder(img)
					b.StartTimer()
					_, err = decoder.DecodeFiles()
					if err != nil {
						b.Fatalf("Error in decoding benchmark: %s", err)
					}
				}
			})
		}
	}
}
