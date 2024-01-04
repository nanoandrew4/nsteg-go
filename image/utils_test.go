package stegimg

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
)

func generateImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	return img
}

func generateImageFile(name string, width, height int) {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: randUint8(), G: randUint8(), B: randUint8(), A: 255})
		}
	}
	imgFile, _ := os.Create(name)
	err := png.Encode(imgFile, img)
	if err != nil {
		panic(err)
	}
}

func randUint8() uint8 {
	return uint8(rand.Intn(256))
}

func generateTestFiles(numOfFilesToGenerate, fileSize int) []string {
	fileNames := make([]string, numOfFilesToGenerate)

	for f := 0; f < numOfFilesToGenerate; f++ {
		fileNames[f] = TestFilePrefix + strconv.Itoa(f)

		generatedBytes := make([]byte, fileSize)
		_, err := rand.Read(generatedBytes)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(fileNames[f], generatedBytes, 0775)
		if err != nil {
			panic(err)
		}
	}

	return fileNames
}

func generateTestFilesFromBytes(numOfFilesToGenerate int, bytesToWrite [][]byte) []string {
	fileNames := make([]string, numOfFilesToGenerate)

	for f := 0; f < numOfFilesToGenerate; f++ {
		fileNames[f] = TestFilePrefix + strconv.Itoa(f)

		err := os.WriteFile(fileNames[f], bytesToWrite[f], 0775)
		if err != nil {
			panic(err)
		}
	}

	return fileNames
}

func chdirToTestFileDir(dir string) {
	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}

	}
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
