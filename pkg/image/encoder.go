package image

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"nsteg/internal/bits"
	"nsteg/pkg/config"
	"nsteg/pkg/model"
	"os"
	"sync"
)

const (
	channelsToWrite = byte(3)
)

var (
	ErrImageNotBigEnough = errors.New("supplied image not big enough to contain the supplied files to hide, either choose another image or increase LSBs to use")
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

type Encoder struct {
	lsbsToUse                         byte
	minChunkSize, chunkSizeMultiplier int
	currentByte, currentPixel         int

	image  *image.RGBA
	config config.ImageEncodeConfig
}

func NewImageEncoder(image *image.RGBA, iConfig config.ImageEncodeConfig) *Encoder {
	iConfig.PopulateUnsetConfigVars()

	enc := &Encoder{
		image:               image,
		config:              iConfig,
		lsbsToUse:           iConfig.LSBsToUse,
		minChunkSize:        int(iConfig.LSBsToUse) * int(channelsToWrite),
		chunkSizeMultiplier: config.DefaultChunkSizeMultiplier,
	}

	enc.encodeLSBsToImage()
	return enc
}

func (ie *Encoder) Encode(dataReader io.Reader, output io.Writer) error {
	ie.encodeDataToRawImage(dataReader)

	return ie.encodeRawImage(output)
}

func (ie *Encoder) EncodeFiles(files []model.InputFile, output io.Writer) error {
	dataToEncode, err := ie.setupDataReader(files)
	if err != nil {
		return err
	}

	ie.encodeDataToRawImage(dataToEncode)
	return ie.encodeRawImage(output)
}

func (ie *Encoder) encodeLSBsToImage() {
	packedLSBsToUse := ie.lsbsToUse - 1 // Save LSBs to use as value 0-7 so it fits in 3 bits (one pixel)
	LSBsBitReader := bits.NewBitReader([]byte{packedLSBsToUse})

	LSBsToUse := ie.lsbsToUse
	for p := 3; p < len(ie.image.Pix); p += 4 {
		if ie.image.Pix[p] == 255 {
			ie.currentPixel = p / 4
			break
		}
	}

	ie.fillPixelLSBs(ie.currentPixel, LSBsBitReader, 1)
	ie.lsbsToUse = LSBsToUse
	ie.currentPixel++
}

func (ie *Encoder) setupDataReader(filesToHide []model.InputFile) (io.Reader, error) {
	var dataReaders []io.Reader

	// Scan ahead to count opaque pixels
	availablePixelChan := make(chan uint64)
	go func() {
		var availablePixels uint64
		for p := 3; p < len(ie.image.Pix); p += 4 {
			if ie.image.Pix[p] == 255 {
				availablePixels++
			}
		}
		availablePixelChan <- availablePixels
	}()

	dataReaders = append(dataReaders, bytes.NewReader(intToBitArray(len(filesToHide))))

	var requiredBitsForEncoding int64
	for _, fileToHide := range filesToHide {
		dataReaders = append(dataReaders,
			bytes.NewReader(intToBitArray(len(fileToHide.Name))),
			bytes.NewReader([]byte(fileToHide.Name)),
			bytes.NewReader(intToBitArray(int(fileToHide.Size))),
			fileToHide.Content)
		requiredBitsForEncoding += (8 + int64(len([]byte(fileToHide.Name))) + 8 + fileToHide.Size) * 8
	}

	if uint64(requiredBitsForEncoding) > <-availablePixelChan*uint64(channelsToWrite) {
		return nil, ErrImageNotBigEnough
	}

	return io.MultiReader(dataReaders...), nil
}

func (ie *Encoder) encodeDataToRawImage(dataReader io.Reader) {
	chunkSize := ie.minChunkSize * ie.chunkSizeMultiplier

	bytesRead := chunkSize
	var eofErr error
	var wg sync.WaitGroup
	for bytesRead == chunkSize && eofErr != io.EOF {
		chunkBytes := make([]byte, chunkSize)
		bytesRead, eofErr = io.ReadFull(dataReader, chunkBytes)
		chunkBytes = chunkBytes[:bytesRead]

		//wg.Add(1)
		//go func(currentPixel int, bytesToWrite []byte) {
		//	defer wg.Done()
		br := bits.NewBitReader(chunkBytes)
		for ; br.BytesLeftToRead() > 0; ie.currentPixel++ {
			ie.fillPixelLSBs(ie.currentPixel, br, ie.lsbsToUse)
		}
		//}(ie.currentPixel, chunkBytes)

		ie.currentByte += chunkSize
		//ie.currentPixel += (chunkSize / int(channelsToWrite) * 8) / int(ie.lsbsToUse)
	}
	wg.Wait()
}

func (ie *Encoder) fillPixelLSBs(pixelToWriteTo int, br *bits.BitReader, LSBsToUse byte) {
	pixelChannelsToOverwrite := ie.image.Pix[pixelToWriteTo*4 : pixelToWriteTo*4+4]
	// Skip non-opaque pixels, since data encoded in them cannot be fully recovered reliably
	if pixelChannelsToOverwrite[3] != 255 {
		return
	}

	// Clear least significant bits to use, and then add the new bits. Iterate backwards since encoding order is green, blue, red, but we need
	// the decoded order to be red, blue, green
	for channel := byte(0); channel < channelsToWrite; channel++ {
		pixelChannelsToOverwrite[channel] = ((pixelChannelsToOverwrite[channel] >> LSBsToUse) << LSBsToUse) + br.ReadBits(uint(LSBsToUse))
	}
}

func (ie *Encoder) encodeRawImage(outputWriter io.Writer) error {
	enc := png.Encoder{CompressionLevel: ie.config.PngCompressionLevel}
	return enc.Encode(outputWriter, ie.image)
}

func (ie *Encoder) saveAsPng(outputPath string, config config.ImageEncodeConfig) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	enc := png.Encoder{CompressionLevel: config.PngCompressionLevel}
	err = enc.Encode(f, ie.image)
	if err != nil {
		return err
	}
	return f.Close()
}

func intToBitArray(i int) []byte {
	byteArr := make([]byte, 8)
	for b := uint(0); b < 8; b++ {
		byteArr[b] = byte((i << (b * 8)) >> 56)
	}

	return byteArr
}
