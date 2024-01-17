package bits

import (
	"fmt"
	"nsteg/test"
	"testing"
)

const numOfBytesForBenchmark = 1000000

func BenchmarkReadBits(b *testing.B) {
	bytesToRead := test.GenerateRandomBytes(numOfBytesForBenchmark)
	for LSBsToUse := uint(1); LSBsToUse <= 8; LSBsToUse++ {
		b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
			b.SetBytes(int64(numOfBytesForBenchmark))
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				bBitReader := NewBitReader(bytesToRead)
				b.StartTimer()
				for len(bBitReader.bytes) > 0 {
					bBitReader.ReadBits(LSBsToUse)
				}
			}
		})
	}
}
