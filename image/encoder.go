package stegimg

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"nsteg/bits"
	"os"
	"strings"
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

func Encode(imageSourcePath, outputPath string, filesToHide []string, config Config) error {
	config.populateUnsetConfigVars()

	srcImage, err := getImageFromFilePath(imageSourcePath)
	if err != nil {
		return err
	}

	encoder := newEncoder(srcImage, config.LSBsToUse)
	filesToHideReader, err := encoder.setupDataReader(filesToHide)
	if err != nil {
		return err
	}
	encoder.encodeDataToImage(filesToHideReader)

	return encoder.saveAsPng(outputPath, config)
}

type imageEncoder struct {
	lsbsToUse                         byte
	minChunkSize, chunkSizeMultiplier int
	currentByte, currentPixel         int

	image *image.RGBA
}

func newEncoder(image *image.RGBA, LSBsToUse byte) *imageEncoder {
	enc := &imageEncoder{
		image:               image,
		lsbsToUse:           LSBsToUse,
		minChunkSize:        int(LSBsToUse) * int(channelsToWrite),
		chunkSizeMultiplier: defaultChunkSizeMultiplier,
	}

	enc.encodeLSBsToImage()
	return enc
}

func (ie *imageEncoder) encodeLSBsToImage() {
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

func (ie *imageEncoder) withChunkSizeMultiplier(chunkSizeMultiplier int) *imageEncoder {
	ie.chunkSizeMultiplier = chunkSizeMultiplier
	return ie
}

func (ie *imageEncoder) setupDataReader(filesToHide []string) (io.Reader, error) {
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
	for f := 0; f < len(filesToHide); f++ {
		filePathSplit := strings.Split(filesToHide[f], "/")
		fileName := filePathSplit[len(filePathSplit)-1]
		dataReaders = append(dataReaders, bytes.NewReader(intToBitArray(len(fileName))))
		dataReaders = append(dataReaders, bytes.NewReader([]byte(fileName)))

		file, err := os.Open(filesToHide[f])
		if err != nil {
			return nil, err
		}
		fileStat, err := file.Stat()
		if err != nil {
			return nil, err
		}
		dataReaders = append(dataReaders, bytes.NewReader(intToBitArray(int(fileStat.Size()))))
		dataReaders = append(dataReaders, file)
		requiredBitsForEncoding += (8 + int64(len([]byte(fileName))) + 8 + fileStat.Size()) * 8
	}

	if uint64(requiredBitsForEncoding) > <-availablePixelChan*uint64(channelsToWrite) {
		return nil, ErrImageNotBigEnough
	}

	return io.MultiReader(dataReaders...), nil
}

func (ie *imageEncoder) encodeDataToImage(dataReader io.Reader) {
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

func (ie *imageEncoder) fillPixelLSBs(pixelToWriteTo int, br *bits.BitReader, LSBsToUse byte) {
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

func (ie *imageEncoder) saveAsPng(outputPath string, config Config) error {
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

func getImageFromFilePath(filePath string) (*image.RGBA, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	srcImage, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	} else if err = f.Close(); err != nil {
		return nil, err
	}

	// TODO: Work with 16-bit images
	img := image.NewRGBA(srcImage.Bounds())
	draw.Draw(img, img.Bounds(), srcImage, img.Bounds().Min, draw.Src)

	return img, nil
}
