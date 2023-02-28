package bits

type BitReader struct {
	bytes      []byte
	currentBit uint
}

func NewBitReader(bytes []byte) *BitReader {
	return &BitReader{
		bytes: bytes,
	}
}

func (br *BitReader) BytesLeftToRead() int {
	return len(br.bytes)
}

// Will likely overwrite all channels in last pixel even if not necessary
func (br *BitReader) ReadBits(bitsToRead uint) (byteWithRequestedBits byte) {
	for bitsRead := uint(0); bitsRead < bitsToRead && len(br.bytes) > 0; {
		if bitsToRead < 8-br.currentBit {
			byteWithRequestedBits += (br.bytes[0] << (8 - br.currentBit - (bitsToRead - bitsRead))) >> (8 - bitsToRead)
			br.currentBit += bitsToRead - bitsRead
			bitsRead += bitsToRead - bitsRead
		} else {
			byteWithRequestedBits += br.bytes[0] >> br.currentBit
			br.bytes = br.bytes[1:]
			bitsRead += 8 - br.currentBit
			br.currentBit = 0
		}
	}
	return byteWithRequestedBits
}
