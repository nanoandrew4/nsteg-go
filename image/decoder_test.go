package stegimg

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func BenchmarkDecodeSpeed(b *testing.B) {
	img := generateImage(10000, 10000)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, bytesToRead := range []int{100000, 1000000, 10000000} {
			numOfBytesToEncode := bytesToRead
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(bytesToRead)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					bytesToEncode := make([]byte, numOfBytesToEncode)
					_, err := rand.Read(bytesToEncode)
					if err != nil {
						panic(err)
					}
					testImageDecoder := imageDecoder{
						image:     img,
						LSBsToUse: LSBsToUse,
					}
					b.StartTimer()
					testImageDecoder.readBytes(bytesToRead)
				}
			})
		}
	}
}