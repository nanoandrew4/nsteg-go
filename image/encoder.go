package stegimg

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
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

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	enc := png.Encoder{CompressionLevel: config.PngCompressionLevel}
	err = enc.Encode(f, encoder.image)
	if err != nil {
		return err
	}
	return f.Close()
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
	packedLSBsToUse := ie.lsbsToUse - 1
	LSBsBitReader := bits.NewBitReader([]byte{packedLSBsToUse})

	LSBsToUse := ie.lsbsToUse
	ie.fillPixelLSBs(ie.currentPixel, LSBsBitReader, 1)
	ie.lsbsToUse = LSBsToUse
	ie.currentPixel = 1
}

func (ie *imageEncoder) withChunkSizeMultiplier(chunkSizeMultiplier int) *imageEncoder {
	ie.chunkSizeMultiplier = chunkSizeMultiplier
	return ie
}

func (ie *imageEncoder) setupDataReader(filesToHide []string) (io.Reader, error) {
	var dataReaders []io.Reader

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

	imageBounds := ie.image.Bounds()
	bitsAvailableForEncoding := uint64(imageBounds.Dx()) * uint64(imageBounds.Dy()) * uint64(ie.lsbsToUse*3)
	if uint64(requiredBitsForEncoding) > bitsAvailableForEncoding {
		log.Fatalf("Image is not large enough - required capacity is %d bytes, but image only has %d with current LSB settings\n", requiredBitsForEncoding/8, bitsAvailableForEncoding/8)
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

		wg.Add(1)
		go func(currentPixel int, bytesToWrite []byte) {
			defer wg.Done()
			br := bits.NewBitReader(chunkBytes)
			for ; br.BytesLeftToRead() > 0; currentPixel++ {
				ie.fillPixelLSBs(currentPixel, br, ie.lsbsToUse)
			}
		}(ie.currentPixel, chunkBytes)

		ie.currentByte += chunkSize
		ie.currentPixel += (chunkSize / 3 * 8) / int(ie.lsbsToUse)
	}
	wg.Wait()
}

func (ie *imageEncoder) fillPixelLSBs(pixelToWrite int, br *bits.BitReader, LSBsToUse byte) {
	pixelChannelsToOverwrite := ie.image.Pix[pixelToWrite*4 : pixelToWrite*4+4]

	// Clear least significant bits to use, and then add the new bits. Iterate backwards since encoding order is green, blue, red, but we need
	// the decoded order to be red, blue, green
	for channel := byte(0); channel < channelsToWrite; channel++ {
		pixelChannelsToOverwrite[channel] = ((pixelChannelsToOverwrite[channel] >> LSBsToUse) << LSBsToUse) + br.ReadBits(uint(LSBsToUse))
	}
}

func intToBitArray(i int) []byte {
	byteArr := make([]byte, 8, 8)
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
