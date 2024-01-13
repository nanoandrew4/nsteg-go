package image

import (
	"errors"
	"image"
	"nsteg/pkg/model"
)

var (
	// TODO: could encode static string as first n bytes for checking
	ErrDecodeFileBounds = errors.New("decoding exceeded image bounds, the file was likely not encoded using nsteg")
)

type Decoder struct {
	LSBsToUse, currentPixelBit byte
	currentPixel               int

	image *image.RGBA
}

func NewImageDecoder(image *image.RGBA) *Decoder {
	d := &Decoder{
		image: image,
	}

	d.decodeLSBsToUse()
	return d
}

func (d *Decoder) DecodeFiles() ([]model.OutputFile, error) {
	var decodedFiles []model.OutputFile

	numOfFilesToDecodeBytes, err := d.readBytes(8)
	if err != nil {
		return nil, err
	}
	for f := 0; f < bytesToInt(numOfFilesToDecodeBytes); f++ {
		fileNameLength, err := d.readBytes(8)
		if err != nil {
			return nil, err
		}
		fileName, err := d.readBytes(bytesToInt(fileNameLength))
		if err != nil {
			return nil, err
		}
		fileLength, err := d.readBytes(8)
		if err != nil {
			return nil, err
		}

		fileBytes, err := d.readBytes(bytesToInt(fileLength))
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

func (d *Decoder) incrementCurrentPixel() {
	d.currentPixel++
	if d.currentPixel%4 == 3 { // Skip alpha channel
		d.currentPixel++
	}
}

func (d *Decoder) decodeLSBsToUse() {
	// Find first opaque pixel, which will contain the LSBs
	for p := 3; p < len(d.image.Pix); p += 4 {
		if d.image.Pix[p] == 255 {
			d.currentPixel = p - 3
			break
		}
	}
	firstPixel := d.image.Pix[d.currentPixel : d.currentPixel+3]
	// Value will be 0-7 (3 bit value), we add 1 to restore the original 1-8 value
	d.LSBsToUse = (firstPixel[0] & 1) + (firstPixel[1]&1)<<1 + (firstPixel[2]&1)<<2 + 1
	d.currentPixel += 4
}

func (d *Decoder) readBytes(numOfBytesToRead int) ([]byte, error) {
	readBytes := make([]byte, numOfBytesToRead)
	var currByte, currBit byte
	var currByteIdx int
	for currByteIdx < numOfBytesToRead {
		if d.currentPixel >= len(d.image.Pix) {
			return nil, ErrDecodeFileBounds
		}

		var currentPixelAlpha = d.image.Pix[(d.currentPixel/4)*4+3]
		if currentPixelAlpha != 255 {
			d.currentPixel += 4
			continue
		}

		if d.currentPixelBit > 0 {
			bitsLeftToReadInPixel := d.LSBsToUse - d.currentPixelBit
			currByte += (d.image.Pix[d.currentPixel] & ((1<<bitsLeftToReadInPixel - 1) << (d.currentPixelBit))) >> d.currentPixelBit
			currBit += bitsLeftToReadInPixel
			d.currentPixelBit = 0
			d.incrementCurrentPixel()
		}

		if currBit+d.LSBsToUse <= 8 {
			currByte += (d.image.Pix[d.currentPixel] & (1<<d.LSBsToUse - 1)) << currBit
			currBit += d.LSBsToUse
			d.incrementCurrentPixel()
		} else {
			if bitsLeftToRead := 8 - currBit; bitsLeftToRead > 0 {
				currByte += (d.image.Pix[d.currentPixel] & (1<<bitsLeftToRead - 1)) << currBit
				d.currentPixelBit = bitsLeftToRead
			}

			readBytes[currByteIdx] = currByte
			currByteIdx++
			currByte = 0
			currBit = 0
		}
	}

	return readBytes, nil
}

func bytesToInt(bytes []byte) int {
	var intFromBytes int
	for i := 0; i < 7; i++ {
		intFromBytes += int(bytes[i])
		intFromBytes <<= 8
	}
	intFromBytes += int(bytes[len(bytes)-1])

	return intFromBytes
}
