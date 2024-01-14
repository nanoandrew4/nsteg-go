package image

import (
	"errors"
	"image"
	"nsteg/pkg/model"
	"time"
)

const (
	MaxBytesAllocatedAtOnce = 1000 * 1000 * 1000
)

var (
	ErrDecodeFileBounds = errors.New("decoding exceeded image bounds, the file was likely not encoded using nsteg")
	ErrMaxAllocExceeded = errors.New("tried to allocate too much memory at once during decoding, which could lead to OOM panic")
)

type Decoder struct {
	LSBsToUse, bitsLeftToReadInSubPixel byte

	// currentSubPixel Represents the pixel/channel the decoder is on. A value of 3, according to the RGBA order of
	// image.RGBA, would represent the blue channel of the first pixel
	currentSubPixel int

	image *image.RGBA
	stats model.DecodeStats
}

func NewImageDecoder(image *image.RGBA) (*Decoder, error) {
	d := &Decoder{
		image: image,
	}

	err := d.decodeLSBsToUse()
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Decoder) Stats() model.DecodeStats {
	return d.stats
}

func (d *Decoder) DecodeFiles() ([]model.OutputFile, error) {
	decodeStart := time.Now()
	defer func() {
		d.stats.DataDecoding = time.Since(decodeStart)
	}()

	var decodedFiles []model.OutputFile

	numOfFilesToDecode, err := d.readUInt()
	if err != nil {
		return nil, err
	}
	for f := uint(0); f < numOfFilesToDecode; f++ {
		fileNameLength, err := d.readUInt()
		if err != nil {
			return nil, err
		}
		fileName, err := d.readBytes(fileNameLength)
		if err != nil {
			return nil, err
		}
		fileLength, err := d.readUInt()
		if err != nil {
			return nil, err
		}
		fileBytes, err := d.readBytes(fileLength)
		if err != nil {
			return nil, err
		}

		decodedFiles = append(decodedFiles, model.OutputFile{
			Name:    string(fileName),
			Content: fileBytes,
		})
	}

	return decodedFiles, nil
}

func (d *Decoder) decodeLSBsToUse() error {
	// Find first opaque pixel, which will contain the LSBs
	var opaquePixelFound bool
	for p := 3; p < len(d.image.Pix); p += 4 {
		if d.image.Pix[p] == 255 {
			d.currentSubPixel = p - 3
			opaquePixelFound = true
			break
		}
	}

	if !opaquePixelFound {
		return ErrDecodeFileBounds
	}

	firstPixel := d.image.Pix[d.currentSubPixel : d.currentSubPixel+3]
	// Value will be 0-7 (3 bit value), we add 1 to restore the original 1-8 value
	d.LSBsToUse = (firstPixel[0] & 1) + (firstPixel[1]&1)<<1 + (firstPixel[2]&1)<<2 + 1
	d.currentSubPixel += 4
	return nil
}

func (d *Decoder) readUInt() (uint, error) {
	intBytes, err := d.readBytes(8)
	if err != nil {
		return 0, err
	}
	return bytesToInt(intBytes), nil
}

func (d *Decoder) readBytes(numOfBytesToRead uint) (b []byte, retErr error) {
	// Images not encoded with nsteg will cause random data to be read, which will likely lead to an attempt to decode a
	// random number of bytes. We set a hard limit to catch these cases and prevent an OOM panic from crashing the
	// program. Although some sort of identifying byte sequence could be encoded to mitigate this issue, it would
	// not prevent it entirely, and would defeat the purpose of steganography, by making it trivial to identify if an
	// image is holding secret data or not
	if numOfBytesToRead > MaxBytesAllocatedAtOnce {
		return nil, ErrMaxAllocExceeded
	}

	d.advanceToNextOpaquePixelIfOnNonOpaquePixel()

	readBytes := make([]byte, numOfBytesToRead)
	var currByte, currBit byte
	var currByteIdx uint
	for currByteIdx < numOfBytesToRead {

		// This path is only traversed by LSB settings of 3, 5, 6 and 7, since these values can cause a pixel to fill a
		// byte, and have some bits leftover to read. As an example, if LSBs == 3, a pixel will contain 9 bits of
		// encoded data, which will fill a byte, and one bit will be left over, this path reads that leftover bit
		// into the current byte
		if d.bitsLeftToReadInSubPixel > 0 {
			bitsPreviouslyReadFromPixel := d.LSBsToUse - d.bitsLeftToReadInSubPixel
			currByte += (d.image.Pix[d.currentSubPixel] & ((1<<d.bitsLeftToReadInSubPixel - 1) << (bitsPreviouslyReadFromPixel))) >> bitsPreviouslyReadFromPixel
			currBit += d.bitsLeftToReadInSubPixel
			d.bitsLeftToReadInSubPixel = 0
			err := d.advanceToNextOpaqueSubpixel()
			if err != nil {
				return nil, err
			}
		}

		// This path is traversed by all LSB settings
		if currBit+d.LSBsToUse <= 8 {
			currByte += (d.image.Pix[d.currentSubPixel] & (1<<d.LSBsToUse - 1)) << currBit
			currBit += d.LSBsToUse
			err := d.advanceToNextOpaqueSubpixel()
			if err != nil {
				return nil, err
			}
		} else { // This path is traversed by all LSB settings

			// This path is only traversed by LSB settings of 3, 5, 6 and 7. As explained above, these LSB settings
			// cause pixels to only be partially read to fill a byte. Here we finish filling the current byte with
			// however bits its missing from the current pixel. The remaining bits will be read on the next iteration
			// in the other if statement targeting these LSB settings
			if bitsReadFromPixel := 8 - currBit; bitsReadFromPixel > 0 {
				currByte += (d.image.Pix[d.currentSubPixel] & (1<<bitsReadFromPixel - 1)) << currBit
				d.bitsLeftToReadInSubPixel = d.LSBsToUse - bitsReadFromPixel
			}

			readBytes[currByteIdx] = currByte
			currByteIdx++
			currByte = 0
			currBit = 0
		}
	}

	return readBytes, nil
}

func (d *Decoder) advanceToNextOpaqueSubpixel() error {
	d.currentSubPixel++
	if d.currentSubPixel%4 == 3 { // Skip alpha channel
		d.currentSubPixel++
		d.advanceToNextOpaquePixelIfOnNonOpaquePixel()
	}

	if d.currentSubPixel >= len(d.image.Pix) {
		return ErrDecodeFileBounds
	}
	return nil
}

func (d *Decoder) advanceToNextOpaquePixelIfOnNonOpaquePixel() {
	for d.image.Pix[(d.currentSubPixel/4)*4+3] != 255 {
		d.currentSubPixel += 4
	}
}

func bytesToInt(bytes []byte) uint {
	var intFromBytes uint
	for i := 0; i < 7; i++ {
		intFromBytes += uint(bytes[i])
		intFromBytes <<= 8
	}
	intFromBytes += uint(bytes[len(bytes)-1])

	return intFromBytes
}
