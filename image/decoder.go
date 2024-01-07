package stegimg

import (
	"image"
	"nsteg/internal/cli"
	"os"
)

func DecodeImg(encodedMediaFile string) error {
	srcImage, err := cli.GetImageFromFilePath(encodedMediaFile)
	if err != nil {
		return err
	}

	decoder := newImageDecoder(srcImage)
	decoder.decodeLSBsToUse()

	numOfFilesToDecode := bytesToInt(decoder.readBytes(8))
	for f := 0; f < numOfFilesToDecode; f++ {
		fileNameLength := bytesToInt(decoder.readBytes(8))
		fileName := string(decoder.readBytes(fileNameLength))
		fileLength := bytesToInt(decoder.readBytes(8))
		fileBytes := decoder.readBytes(fileLength)
		err = os.WriteFile(fileName, fileBytes, 0664)
		if err != nil {
			return err
		}
	}
	return nil
}

type imageDecoder struct {
	LSBsToUse, currentPixelBit byte
	currentPixel               int

	image *image.RGBA
}

func newImageDecoder(image *image.RGBA) *imageDecoder {
	return &imageDecoder{
		image: image,
	}
}

func (d *imageDecoder) incrementCurrentPixel() {
	d.currentPixel++
	if d.currentPixel%4 == 3 { // Skip alpha channel
		d.currentPixel++
	}
}

func (d *imageDecoder) decodeLSBsToUse() {
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

func (d *imageDecoder) readBytes(numOfBytesToRead int) []byte {
	readBytes := make([]byte, numOfBytesToRead)
	var currByte, currBit byte
	var currByteIdx int
	for currByteIdx < numOfBytesToRead {
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

	return readBytes
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
