package image

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image/png"
	"io"
	"nsteg/pkg/config"
	"testing"
)

func BenchmarkEncodeSpeedOnOpaqueImage(b *testing.B) {
	compressionLevelNames := map[png.CompressionLevel]string{
		png.NoCompression:      "none",
		png.DefaultCompression: "default",
		png.BestSpeed:          "fast",
		png.BestCompression:    "best",
	}

	img, _ := generateImage(10000, 10000, false)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, byteSize := range []int{100000, 1000000, 10000000} {
			for _, compressionLevel := range []png.CompressionLevel{png.NoCompression, png.DefaultCompression, png.BestSpeed, png.BestCompression} {
				numOfBytesToEncode := byteSize
				b.Run(fmt.Sprintf("LSBsToUse=%d,png.CompressionLevel=%s,MBs=%f",
					LSBsToUse, compressionLevelNames[compressionLevel], float64(byteSize)/1000000.0), func(b *testing.B) {
					b.SetBytes(int64(numOfBytesToEncode))
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						bytesToEncode := make([]byte, numOfBytesToEncode)
						_, err := rand.Read(bytesToEncode)
						if err != nil {
							panic(err)
						}
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
						_ = testImageEncoder.Encode(bytesReader, io.Discard)
					}
				})
			}
		}
	}
}
