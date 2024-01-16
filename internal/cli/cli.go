package cli

import (
	"fmt"
	"image"
	"image/draw"
	"nsteg/pkg/config"
	nstegImage "nsteg/pkg/image"
	"nsteg/pkg/model"
	"os"
)

func EncodeImageWithFiles(imageSourcePath, outputPath string, fileNames []string, config config.ImageEncodeConfig) error {
	srcImage, err := GetImageFromFilePath(imageSourcePath)
	if err != nil {
		return err
	}

	iEncoder, err := nstegImage.NewImageEncoder(srcImage, config)
	if err != nil {
		return err
	}

	var filesToHide []model.InputFile
	for _, fileName := range fileNames {
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		fileStat, err := file.Stat()
		if err != nil {
			return err
		}
		filesToHide = append(filesToHide, model.InputFile{
			Name:    file.Name(),
			Content: file,
			Size:    fileStat.Size(),
		})
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer outputFile.Close()
	err = iEncoder.EncodeFiles(filesToHide)
	if err != nil {
		return err
	}
	err = iEncoder.WriteEncodedPNG(outputFile)
	if err != nil {
		return err
	}

	fmt.Printf("Encoder setup time: %s\n", iEncoder.Stats().Setup)
	fmt.Printf("Data encode time: %s\n", iEncoder.Stats().DataEncoding)
	fmt.Printf("Output image encode time: %s\n", iEncoder.Stats().OutputImageEncoding)
	return nil
}

func DecodeFilesFromImage(encodedMediaFile string) error {
	srcImage, err := GetImageFromFilePath(encodedMediaFile)
	if err != nil {
		return err
	}

	decoder, err := nstegImage.NewImageDecoder(srcImage)
	if err != nil {
		return err
	}

	decodedFiles, err := decoder.DecodeFiles()
	if err != nil {
		return err
	}
	for _, decodedFile := range decodedFiles {
		err = os.WriteFile(decodedFile.Name, decodedFile.Content, 0664)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetImageFromFilePath(filePath string) (*image.RGBA, error) {
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
