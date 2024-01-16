package bits

// BitReader implements methods to help with reading bits from an array of bytes. Bits are read from least significant
// to most significant
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

func (br *BitReader) BitsLeftToRead() int {
	return (len(br.bytes)-1)*8 + (8 - int(br.currentBitIdx))
}

func (br *BitReader) Reset() {
	br.bytes = nil
	br.currentBitIdx = 0
}

// Will likely overwrite all channels in last pixel even if not necessary
func (br *BitReader) ReadBits(bitsToRead uint) (byteWithRequestedBits byte) {
	var numOfBitsRead uint
	for numOfBitsRead < bitsToRead && len(br.bytes) > 0 {
		if bitsToRead < 8-br.currentBitIdx {
			byteWithRequestedBits += (br.bytes[0] << (8 - br.currentBitIdx - (bitsToRead - numOfBitsRead))) >> (8 - bitsToRead)
			br.currentBitIdx += bitsToRead - numOfBitsRead
			numOfBitsRead += bitsToRead - numOfBitsRead
		} else {
			byteWithRequestedBits += br.bytes[0] >> br.currentBitIdx
			br.bytes = br.bytes[1:]
			numOfBitsRead += 8 - br.currentBitIdx
			br.currentBitIdx = 0
		}
	}
	return byteWithRequestedBits
}
