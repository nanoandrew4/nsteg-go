package image

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"nsteg/pkg/config"
	"nsteg/test"
	"testing"
)

const (
	benchImageSize = 5000
)

func BenchmarkEncodeWithPNGOutput(b *testing.B) {
	compressionLevelNames := map[png.CompressionLevel]string{
		png.NoCompression:      "none",
		png.DefaultCompression: "default",
		png.BestSpeed:          "fast",
		png.BestCompression:    "best",
	}

	compressionLevelsToBenchmark := []png.CompressionLevel{
		png.NoCompression, png.DefaultCompression, png.BestSpeed, png.BestCompression,
	}

	for _, randomizePixelOpaqueness := range []bool{false, true} {
		b.Run(getOpaquenessLabel(randomizePixelOpaqueness), func(b *testing.B) {
			img, opaquePixels := generateImage(benchImageSize, benchImageSize, randomizePixelOpaqueness)
			for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
				numOfBytesToEncode := calculateBytesThatFitInImage(opaquePixels, LSBsToUse)
				for _, compressionLevel := range compressionLevelsToBenchmark {
					benchmarkLabel := fmt.Sprintf("LSBsToUse=%d,png.CompressionLevel=%s", LSBsToUse,
						compressionLevelNames[compressionLevel])

					b.Run(benchmarkLabel, func(b *testing.B) {
						b.SetBytes(int64(numOfBytesToEncode))
						for i := 0; i < b.N; i++ {
							b.StopTimer()
							bytesToEncode := test.GenerateRandomBytes(numOfBytesToEncode)
							iConfig := config.ImageEncodeConfig{
								LSBsToUse:           LSBsToUse,
								PngCompressionLevel: compressionLevel,
							}
							testImageEncoder, err := NewImageEncoder(img, iConfig)
							if err != nil {
								b.Fatalf("Error creating image encoder for benchmark")
							}
							bytesReader := bytes.NewReader(bytesToEncode)
							b.StartTimer()
							err = testImageEncoder.Encode(bytesReader)
							if err != nil {
								b.Fatalf("Error during image encoding: %s", err)
							}
							err = testImageEncoder.WriteEncodedPNG(io.Discard)
							if err != nil {
								b.Fatalf("Error writing PNG image: %s", err)
							}
						}
					})
				}
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, randomizePixelOpaqueness := range []bool{false, true} {
		b.Run(getOpaquenessLabel(randomizePixelOpaqueness), func(b *testing.B) {
			img, opaquePixels := generateImage(benchImageSize, benchImageSize, randomizePixelOpaqueness)
			for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
				numOfBytesToEncode := calculateBytesThatFitInImage(opaquePixels, LSBsToUse)
				b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
					b.SetBytes(int64(numOfBytesToEncode))
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						bytesToEncode := test.GenerateRandomBytes(numOfBytesToEncode)
						iConfig := config.ImageEncodeConfig{
							LSBsToUse: LSBsToUse,
						}
						testImageEncoder, err := NewImageEncoder(img, iConfig)
						if err != nil {
							b.Fatalf("Error creating image encoder for benchmark")
						}
						bytesReader := bytes.NewReader(bytesToEncode)
						b.StartTimer()
						err = testImageEncoder.Encode(bytesReader)
						if err != nil {
							b.Fatalf("Error during image encoding: %s", err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkEncodeFiles(b *testing.B) {
	for _, randomizePixelOpaqueness := range []bool{false, true} {
		b.Run(getOpaquenessLabel(randomizePixelOpaqueness), func(b *testing.B) {
			img, opaquePixels := generateImage(benchImageSize, benchImageSize, randomizePixelOpaqueness)
			for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
				numOfBytesToGenerate := calculateBytesThatFitInImage(opaquePixels, LSBsToUse)
				testFiles := generateFilesToEncode(numOfBytesToGenerate)
				b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
					b.SetBytes(int64(numOfBytesToGenerate))
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						iConfig := config.ImageEncodeConfig{
							LSBsToUse: LSBsToUse,
						}
						testImageEncoder, err := NewImageEncoder(img, iConfig)
						if err != nil {
							b.Fatalf("Error creating image encoder for benchmark")
						}
						filesToEncode := convertTestInputToStandardInput(testFiles)
						b.StartTimer()
						err = testImageEncoder.EncodeFiles(filesToEncode)
						if err != nil {
							b.Fatalf("Error during image encoding: %s", err)
						}
					}
				})
			}
		})
	}
}
