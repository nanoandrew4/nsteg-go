package stegimg

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func BenchmarkReadBits(b *testing.B) {
	for LSBsToUse := uint(1); LSBsToUse <= 8; LSBsToUse++ {
		initBitsFromByteMap(byte(LSBsToUse))
		bytesForBenchmark := make([]byte, 1000000000)
		_, err := rand.Read(bytesForBenchmark)
		if err != nil {
			panic(err)
		}
		bBitReader := newBitReader(bytesForBenchmark)

		b.Run(fmt.Sprintf("LSBsToUse=%d", LSBsToUse), func(b *testing.B) {
			for i := 0; len(bBitReader.bytes) > 0 && i < b.N; i++ {
				bBitReader.readBits(LSBsToUse)
			}
		})
	}
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
		initBitsFromByteMap(byte(bitsToRead))
		tBitReader := newBitReader(bytesToTestWith)
		for iter, expectedBits := range expectedBitsToRead[bitsToRead-1] {
			bits := tBitReader.readBits(bitsToRead)
			if bits != expectedBits {
				t.Errorf("Failure testing bit reader with %d bits per read on iter %d, result was: %d, expected %d", bitsToRead, iter+1, bits, expectedBits)
			}
		}
	}
}
