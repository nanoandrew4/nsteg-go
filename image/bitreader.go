package stegimg

import "errors"

type bitReader struct {
	bytes      []byte
	currentBit uint
}

func newBitReader(bytes []byte) *bitReader {
	return &bitReader{
		bytes: bytes,
	}
}

var bitsFromByteMap map[byte][8]byte

func initBitsFromByteMap(LSBsToUse byte) {
	bitsFromByteMap = map[byte][8]byte{}

	for b := byte(0); ; b++ {
		byteBits := [8]byte{}
		for bitPos := byte(0); bitPos < 8; bitPos++ {
			mask := byte(1<<(LSBsToUse+bitPos)) - 1
			if bitPos > 0 {
				mask -= 1 << (bitPos - 1)
			}
			byteBits[bitPos] = (b & mask) >> bitPos
		}
		bitsFromByteMap[b] = byteBits

		if b == byte(255) {
			break
		}
	}
}

// Will likely overwrite all channels in last pixel even if not necessary
func (br *bitReader) readBits(bitsToRead uint) byte {
	if bitsFromByteMap == nil {
		panic(errors.New("bitsFromByteMap is nil"))
	}

	var byteWithRequestedBits byte
	for bitsRead := uint(0); bitsRead < bitsToRead && len(br.bytes) > 0; {
		byteWithRequestedBits += (bitsFromByteMap[br.bytes[0]][br.currentBit] & byte(1<<(bitsToRead-bitsRead)-1)) << bitsRead
		if bitsToRead < 8-br.currentBit {
			br.currentBit += bitsToRead - bitsRead
			bitsRead += bitsToRead - bitsRead
		} else {
			br.bytes = br.bytes[1:]
			bitsRead += 8 - br.currentBit
			br.currentBit = 0
		}
	}
	return byteWithRequestedBits
}
