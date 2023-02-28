package bits

import (
	"crypto/rand"
	"fmt"
	"testing"
)

const numOfBytesForBenchmark = 1000000

func BenchmarkReadBits(b *testing.B) {
	var bBitReader *BitReader
	for LSBsToUse := uint(1); LSBsToUse <= 8; LSBsToUse++ {
		b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
			var bytesRead int
			for i := 0; i < b.N; {
				b.StopTimer()
				bBitReader = initBenchmarkBitReader()
				b.StartTimer()
				for ; i < b.N && len(bBitReader.bytes) > 0; i++ {
					bBitReader.ReadBits(LSBsToUse)
				}
				b.StopTimer()
				if len(bBitReader.bytes) == 0 {
					bytesRead += numOfBytesForBenchmark
				}
			}
			bytesRead += numOfBytesForBenchmark - len(bBitReader.bytes)
			b.SetBytes(int64(bytesRead))
		})
	}
}

func initBenchmarkBitReader() *BitReader {
	bytesForBenchmark := make([]byte, numOfBytesForBenchmark)
	_, err := rand.Read(bytesForBenchmark)
	if err != nil {
		panic(err)
	}
	return NewBitReader(bytesForBenchmark)
}

func TestReadBits(t *testing.T) {

	// 10000000 00000111 11111111 01100101
	bytesToTestWith := []byte{128, 7, 255, 101}
	expectedBitsToRead := [8][]byte{
		{0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 1},
		{0, 0, 0, 2, 3, 1, 0, 0, 3, 3, 3, 3, 1, 1, 2, 1},
		{0, 0, 6, 3, 0, 6, 7, 7, 5, 4, 1},
		{0, 8, 7, 0, 15, 15, 5, 6},
		{0, 28, 1, 30, 31, 18, 1},
		{0, 30, 48, 63, 37, 1},
		{0, 15, 124, 47, 6},
		{128, 7, 255, 101},
	}

	for bitsToRead := uint(1); bitsToRead <= 8; bitsToRead++ {
		tBitReader := NewBitReader(bytesToTestWith)
		for iter, expectedBits := range expectedBitsToRead[bitsToRead-1] {
			bits := tBitReader.ReadBits(bitsToRead)
			if bits != expectedBits {
				t.Errorf("Failure testing bit reader with %d bits per read on iter %d, result was: %d, expected %d", bitsToRead, iter+1, bits, expectedBits)
			}
		}
	}
}
