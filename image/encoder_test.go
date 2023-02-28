package stegimg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"nsteg/bits"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	encTestFolder   = "enctest"
	encTestImage    = "enctest.png"
	encTestOutImage = "enctestout.png"
)

func TestEncode(t *testing.T) {
	chdirToTestFileDir(encTestFolder)

	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		generateImageFile(encTestImage, 100, 100)

		testFiles := generateTestFiles(3, 1024)
		Encode(encTestImage, encTestOutImage, testFiles, Config{LSBsToUse: LSBsToUse})

		outputImage, err := getImageFromFilePath(encTestOutImage)
		if err != nil {
			t.Fatalf("Error reading output image %s", encTestOutImage)
		}

		firstPixel := outputImage.Pix[0:4]
		encodedLSBsToUse := (firstPixel[0] & 1) + (firstPixel[1]&1)<<1 + (firstPixel[2]&1)<<2 + 1
		if LSBsToUse != encodedLSBsToUse {
			t.Errorf("Expected encoded LSBsToUse to be %d, was %d", LSBsToUse, encodedLSBsToUse)
		}

		var expectedEncodedBytes []byte
		expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(len(testFiles))...)
		for f := 0; f < len(testFiles); f++ {
			filePathSplit := strings.Split(testFiles[f], "/")
			fileName := filePathSplit[len(filePathSplit)-1]
			expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(len(fileName))...)
			expectedEncodedBytes = append(expectedEncodedBytes, fileName...)

			file, err := os.Open(testFiles[f])
			if err != nil {
				log.Fatalf("Error opening file %s", testFiles[f])
			}
			fileStat, err := file.Stat()
			if err != nil {
				log.Fatalf("Error opening file info for file %s", testFiles[f])
			}
			expectedEncodedBytes = append(expectedEncodedBytes, intToBitArray(int(fileStat.Size()))...)
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				log.Fatalf("Error reading file %s", fileName)
			}
			expectedEncodedBytes = append(expectedEncodedBytes, fileBytes...)
		}

		testBitReader := bits.NewBitReader(expectedEncodedBytes)
		for currentPixel := 1; testBitReader.BytesLeftToRead() > 0; currentPixel++ {
			px := currentPixel % outputImage.Bounds().Dx()
			py := currentPixel / outputImage.Bounds().Dx()
			pixelOffset := outputImage.PixOffset(px, py)
			pixel := outputImage.Pix[pixelOffset : pixelOffset+4]
			for channelIdx := byte(0); channelIdx < channelsToWrite; channelIdx++ {
				bitsToCheck := pixel[channelIdx] & (1<<LSBsToUse - 1)
				expectedBits := testBitReader.ReadBits(uint(LSBsToUse))
				if bitsToCheck != expectedBits {
					t.Fatalf("Error with LSBs %d in pixel|channel %d|%d, expected|got %d|%d", LSBsToUse, currentPixel+1, channelIdx+1, expectedBits, bitsToCheck)
				}
			}
		}

		deleteFiles(testFiles)
	}

	os.Chdir("..")
	os.RemoveAll(encTestFolder)
}

func BenchmarkEncodeSpeed(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	img := generateImage(10000, 10000)
	for LSBsToUse := byte(1); LSBsToUse <= 8; LSBsToUse++ {
		for _, byteSize := range []int{100000, 1000000, 10000000} {
			numOfBytesToEncode := byteSize
			b.Run(fmt.Sprintf("MBs=%f/LSBsToUse=%d", float64(byteSize)/1000000.0, LSBsToUse), func(b *testing.B) {
				b.SetBytes(int64(numOfBytesToEncode))
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					bytesToEncode := make([]byte, numOfBytesToEncode)
					_, err := rand.Read(bytesToEncode)
					if err != nil {
						panic(err)
					}
					testImageEncoder := newEncoder(img, LSBsToUse)
					bytesReader := bytes.NewReader(bytesToEncode)
					b.StartTimer()
					testImageEncoder.encodeDataToImage(bytesReader)
				}
			})
		}
	}
}
