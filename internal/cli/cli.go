package cli

import (
	"image"
	"image/draw"
	"nsteg/internal"
	"nsteg/internal/encoder"
	"os"
)

func EncodeImageWithFiles(imageSourcePath, outputPath string, fileNames []string, config internal.ImageEncodeConfig) error {
	srcImage, err := GetImageFromFilePath(imageSourcePath)
	if err != nil {
		return err
	}

	iEncoder := encoder.NewImageEncoder(srcImage, config)

	var filesToHide []internal.FileToHide
	for _, fileName := range fileNames {
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		fileStat, err := file.Stat()
		if err != nil {
			return err
		}
		filesToHide = append(filesToHide, internal.FileToHide{
			Name:    file.Name(),
			Size:    fileStat.Size(),
			Content: file,
		})
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer outputFile.Close()
	return iEncoder.EncodeFiles(filesToHide, outputFile)
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
