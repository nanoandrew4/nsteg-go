package cli

import (
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

	iEncoder := nstegImage.NewImageEncoder(srcImage, config)

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
	return iEncoder.EncodeFiles(filesToHide, outputFile)
}

func DecodeFilesFromImage(encodedMediaFile string) error {
	srcImage, err := GetImageFromFilePath(encodedMediaFile)
	if err != nil {
		return err
	}

	decoder := nstegImage.NewImageDecoder(srcImage)

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
