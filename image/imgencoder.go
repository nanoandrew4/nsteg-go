package stegimg

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	channelsToWrite = byte(3)
)

var (
	chunkSizeMultiplier = 128 * 1024
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)
}

func EncodeImg(imgSource string, outputPath string, filesToHide []string, LSBsToUse byte) {
	srcImage, err := getImageFromFilePath(imgSource)
	bitsAvailableForEncoding := uint64(srcImage.Bounds().Dx()) * uint64(srcImage.Bounds().Dy()) * uint64(LSBsToUse*3)
	if err != nil {
		fmt.Fprint(os.Stderr, "An error ocurred when opening image '", imgSource, "':", err)
	}

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
			log.Fatalf("Error opening file %s", filesToHide[f])
		}
		fileStat, err := file.Stat()
		if err != nil {
			log.Fatalf("Error opening file info for file %s", filesToHide[f])
		}
		dataReaders = append(dataReaders, bytes.NewReader(intToBitArray(int(fileStat.Size()))))
		dataReaders = append(dataReaders, file)
		requiredBitsForEncoding += (8 + int64(len([]byte(fileName))) + 8 + fileStat.Size()) * 8
	}

	if uint64(requiredBitsForEncoding) > bitsAvailableForEncoding {
		log.Fatalf("Image is not large enough - required capacity is %d bytes, but image only has %d with current LSB settings\n", requiredBitsForEncoding/8, bitsAvailableForEncoding/8)
		return
	}

	mReader := io.MultiReader(dataReaders...)

	initBitsFromByteMap(LSBsToUse)
	imgEncoder := newImageEncoder(srcImage, LSBsToUse)
	imgEncoder.encodeLSBsToImage()
	imgEncoder.encodeDataToImage(mReader)

	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatal("Error creating output file")
	}
	defer f.Close()

	png.Encode(f, srcImage)
}

type imageEncoder struct {
	LSBsToUse                 byte
	minChunkSize              int
	currentByte, currentPixel int

	image *image.RGBA
}

func newImageEncoder(image *image.RGBA, LSBsToUse byte) *imageEncoder {
	return &imageEncoder{
		image:        image,
		LSBsToUse:    LSBsToUse,
		minChunkSize: int(LSBsToUse) * int(channelsToWrite),
	}
}

func (ie *imageEncoder) encodeLSBsToImage() {
	packedLSBsToUse := ie.LSBsToUse - 1
	LSBsBitReader := newBitReader([]byte{packedLSBsToUse})

	LSBsToUse := ie.LSBsToUse
	ie.LSBsToUse = 1
	ie.fillPixelLSBs(ie.currentPixel, LSBsBitReader)
	ie.LSBsToUse = LSBsToUse
	ie.currentPixel = 1
}

func (ie *imageEncoder) encodeDataToImage(dataReader io.Reader) {
	chunkSize := ie.minChunkSize * chunkSizeMultiplier

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
			br := newBitReader(chunkBytes)
			for ; len(br.bytes) > 0; currentPixel++ {
				ie.fillPixelLSBs(currentPixel, br)
			}
		}(ie.currentPixel, chunkBytes)

		ie.currentByte += chunkSize
		ie.currentPixel += (chunkSize / 3 * 8) / int(ie.LSBsToUse)
	}
	wg.Wait()
}

func (ie *imageEncoder) fillPixelLSBs(pixelToWrite int, br *bitReader) {
	pixel := ie.image.Pix[pixelToWrite*4 : pixelToWrite*4+4]

	// Clear least significant bits to use, and then add the new bits. Iterate backwards since encoding order is green, blue, red, but we need
	// the decoded order to be red, blue, green
	for channel := byte(0); channel < channelsToWrite; channel++ {
		pixel[channel] = ((pixel[channel] >> ie.LSBsToUse) << ie.LSBsToUse) + br.readBits(uint(ie.LSBsToUse))
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
	defer f.Close()
	srcImage, _, err := image.Decode(f)

	// TODO: Work with 16-bit images
	img := image.NewRGBA(srcImage.Bounds())
	draw.Draw(img, img.Bounds(), srcImage, img.Bounds().Min, draw.Src)

	return img, err
}
