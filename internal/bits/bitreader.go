package bits

type BitReader struct {
	bytes         []byte
	currentBitIdx uint
}

func NewBitReader(bytes []byte) *BitReader {
	return &BitReader{
		bytes: bytes,
	}
}

func (br *BitReader) BytesLeftToRead() int {
	return len(br.bytes)
}

func (br *BitReader) Reset() {
	br.bytes = nil
	br.currentBitIdx = 0
}

// Will likely overwrite all channels in last pixel even if not necessary
func (br *BitReader) ReadBits(bitsToRead uint) (byteWithRequestedBits byte) {
	for bitsRead := uint(0); bitsRead < bitsToRead && len(br.bytes) > 0; {
		if bitsToRead < 8-br.currentBitIdx {
			byteWithRequestedBits += (br.bytes[0] << (8 - br.currentBitIdx - (bitsToRead - bitsRead))) >> (8 - bitsToRead)
			br.currentBitIdx += bitsToRead - bitsRead
			bitsRead += bitsToRead - bitsRead
		} else {
			byteWithRequestedBits += br.bytes[0] >> br.currentBitIdx
			br.bytes = br.bytes[1:]
			bitsRead += 8 - br.currentBitIdx
			br.currentBitIdx = 0
		}
	}
	return byteWithRequestedBits
}
