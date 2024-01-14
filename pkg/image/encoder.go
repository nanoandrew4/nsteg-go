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
	"sync"
	"time"
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
	minChunkSize, chunkSizeMultiplier int
	currentByte, currentSubPixel      int

	image  *image.RGBA
	config config.ImageEncodeConfig
	stats  model.EncodeStats
}

func NewImageEncoder(image *image.RGBA, iConfig config.ImageEncodeConfig) (*Encoder, error) {
	iConfig.PopulateUnsetConfigVars()

	enc := &Encoder{
		image:               image,
		config:              iConfig,
		minChunkSize:        int(iConfig.LSBsToUse) * int(channelsToWrite),
		chunkSizeMultiplier: config.DefaultChunkSizeMultiplier,
	}

	err := enc.encodeLSBsToImage()
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func (e *Encoder) Stats() model.EncodeStats {
	return e.stats
}

func (e *Encoder) Encode(dataReader io.Reader, output io.Writer) error {
	e.encodeDataToRawImage(dataReader)

	return e.encodeRawImage(output)
}

func (e *Encoder) EncodeFiles(files []model.InputFile, output io.Writer) error {
	e.stats = model.EncodeStats{}

	dataToEncode, err := e.setupDataReader(files)
	if err != nil {
		return err
	}

	e.encodeDataToRawImage(dataToEncode)
	return e.encodeRawImage(output)
}

func (e *Encoder) encodeLSBsToImage() error {
	packedLSBsToUse := e.config.LSBsToUse - 1 // Save LSBs to use as value 0-7 so it fits in 3 bits (one pixel)
	LSBsBitReader := bits.NewBitReader([]byte{packedLSBsToUse})

	var opaquePixelFound bool
	for p := 3; p < len(e.image.Pix); p += 4 {
		if e.image.Pix[p] == 255 {
			e.currentSubPixel = p - 3
			opaquePixelFound = true
			break
		}
	}

	if !opaquePixelFound {
		return ErrImageNotBigEnough
	}

	for i := 0; i < 3; i++ {
		e.fillSubPixelLSBs(LSBsBitReader, 1)
		e.currentSubPixel++
	}
	e.currentSubPixel++
	return nil
}

func (e *Encoder) setupDataReader(filesToHide []model.InputFile) (io.Reader, error) {
	setupStart := time.Now()
	defer func() {
		e.stats.Setup = time.Since(setupStart)
	}()

	var dataReaders []io.Reader

	// Scan ahead to count opaque pixels
	availablePixelChan := make(chan uint64)
	go func() {
		var availablePixels uint64
		for p := 3; p < len(e.image.Pix); p += 4 {
			if e.image.Pix[p] == 255 {
				availablePixels++
			}
		}
		availablePixelChan <- availablePixels
	}()

	dataReaders = append(dataReaders, bytes.NewReader(intToBitArray(len(filesToHide))))

	// encoding requires 3 bits for the LSBs setting and 64 (8 bytes) for the number of files encoded aside from,
	// aside from the length of the encoded data
	requiredBitsForEncoding := int64(3 + 64)
	for _, fileToHide := range filesToHide {
		dataReaders = append(dataReaders,
			bytes.NewReader(intToBitArray(len(fileToHide.Name))),
			bytes.NewReader([]byte(fileToHide.Name)),
			bytes.NewReader(intToBitArray(int(fileToHide.Size))),
			fileToHide.Content)

		// length of file name (8 bytes) + file name + length of file (8 bytes) + file contents
		requiredBitsForEncoding += (8 + int64(len([]byte(fileToHide.Name))) + 8 + fileToHide.Size) * 8
	}

	if uint64(requiredBitsForEncoding) > <-availablePixelChan*uint64(channelsToWrite)*uint64(e.config.LSBsToUse) {
		return nil, ErrImageNotBigEnough
	}

	return io.MultiReader(dataReaders...), nil
}

func (e *Encoder) encodeDataToRawImage(dataReader io.Reader) {
	encodeStart := time.Now()
	defer func() {
		e.stats.DataEncoding = time.Since(encodeStart)
	}()

	chunkSize := e.minChunkSize * e.chunkSizeMultiplier

	bytesRead := chunkSize
	var eofErr error
	var wg sync.WaitGroup
	for bytesRead == chunkSize && eofErr != io.EOF {
		chunkBytes := make([]byte, chunkSize)
		bytesRead, eofErr = io.ReadFull(dataReader, chunkBytes)
		chunkBytes = chunkBytes[:bytesRead]

		//wg.Add(1)
		//go func(currentSubPixel int, bytesToWrite []byte) {
		//	defer wg.Done()
		br := bits.NewBitReader(chunkBytes)
		for br.BytesLeftToRead() > 0 && e.currentSubPixel < len(e.image.Pix) {
			subPixelInCurrentPixel := e.currentSubPixel % 4
			if subPixelInCurrentPixel == 0 && e.image.Pix[e.currentSubPixel+3] != 255 {
				e.currentSubPixel += 4 // Skip to next pixel, since data encoded in non-opaque pixels cannot be recovered reliably
			} else {
				if subPixelInCurrentPixel != 3 {
					e.fillSubPixelLSBs(br, e.config.LSBsToUse)
				}
				e.currentSubPixel++
			}
		}
		//}(e.currentSubPixel, chunkBytes)

		e.currentByte += chunkSize
		//e.currentSubPixel += (chunkSize / int(channelsToWrite) * 8) / int(e.lsbsToUse)
	}
	wg.Wait()
}

func (e *Encoder) fillSubPixelLSBs(br *bits.BitReader, LSBsToUse byte) {
	// Clear least significant bits to use, and then add the new bits
	e.image.Pix[e.currentSubPixel] = ((e.image.Pix[e.currentSubPixel] >> LSBsToUse) << LSBsToUse) + br.ReadBits(uint(LSBsToUse))
}

func (e *Encoder) encodeRawImage(outputWriter io.Writer) error {
	imageEncodeStart := time.Now()
	defer func() {
		e.stats.OutputImageEncoding = time.Since(imageEncodeStart)
	}()
	enc := png.Encoder{CompressionLevel: e.config.PngCompressionLevel}
	return enc.Encode(outputWriter, e.image)
}

func intToBitArray(i int) []byte {
	byteArr := make([]byte, 8)
	for b := uint(0); b < 8; b++ {
		byteArr[b] = byte((i << (b * 8)) >> 56)
	}

	return byteArr
}
