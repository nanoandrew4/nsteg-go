package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"image"
	"image/draw"
	"image/png"
	"nsteg/pkg/config"
	nstegImage "nsteg/pkg/image"
	"nsteg/pkg/model"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	pngCompressionMapping = map[string]png.CompressionLevel{
		"default": 0,
		"none":    -1,
		"fast":    -2,
		"best":    -3,
	}
)

func ImageCommands() *cobra.Command {
	imageCmd := &cobra.Command{
		Use:     "image",
		Short:   "Performs steganography operations on images",
		Example: "nsteg image encode --image source.png --output-file output.png --files file1.txt,file2.txt --files file3.txt",
	}

	imageCmd.AddCommand(encodeImageCommand(), decodeFilesFromImage())
	return imageCmd
}

type commonOpts struct {
	lsbsToUse           int8
	chunkSizeMultiplier int
	pngCompression      string
}

func (o commonOpts) toEncodeConfig() config.ImageEncodeConfig {
	mappedCompression, found := pngCompressionMapping[o.pngCompression]
	if !found {
		mappedCompression = png.DefaultCompression
	}
	return config.ImageEncodeConfig{
		LSBsToUse:           byte(o.lsbsToUse),
		ChunkSizeMultiplier: o.chunkSizeMultiplier,
		PngCompressionLevel: mappedCompression,
	}
}

type encodeImageOpts struct {
	sourceImage string
	outputImage string
	fileNames   []string
	config      commonOpts
}

func encodeImageCommand() *cobra.Command {
	opts := encodeImageOpts{}

	encImgCmd := &cobra.Command{
		Use:     "encode",
		Example: "nsteg image encode --image source.png --output-file output.png --files file1.txt,file2.txt --files file3.txt",
		Short:   "Encode data into an image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return EncodeImageWithFiles(opts.sourceImage, opts.outputImage, opts.fileNames, opts.config.toEncodeConfig())
		},
	}

	encImgCmd.Flags().StringVar(&opts.sourceImage, "image", "", "Image to encode data to")
	encImgCmd.Flags().StringVar(&opts.outputImage, "output-file", "", "Name for the encoded image that will be generated")
	encImgCmd.Flags().StringSliceVar(&opts.fileNames, "files", nil, "Files to encode into the source image. Can be comma separated, or you can supply the files param several times with each file")

	encImgCmd.Flags().Int8Var(&opts.config.lsbsToUse, "lsbs", 3, "Least significant bits to use from each pixel. Can be 1-8. The more LSBs are used, the more distortion will be noticeable in the final image")
	encImgCmd.Flags().IntVar(&opts.config.chunkSizeMultiplier, "chunk-size-multiplier", config.DefaultChunkSizeMultiplier, "Chunk size to be handled by a single goroutine")
	encImgCmd.Flags().StringVar(&opts.config.pngCompression, "png-compression", "default", "Compression for output png. Options are default, none, fast, best")

	MarkFlagsRequired(encImgCmd, "image", "output-file", "files")

	return encImgCmd
}

func EncodeImageWithFiles(imageSourcePath, outputPath string, fileNames []string, config config.ImageEncodeConfig) error {
	srcImage, err := getImageFromFilePath(imageSourcePath)
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := NewSpinner()
		s.FinalMSG = fmt.Sprintf("Generated %s which has the following files encoded: %s\n", outputPath, strings.Join(fileNames, ","))
		s.Start()
		for {
			if iEncoder.Stats().Setup == 0 {
				s.Prefix = "Setting up encoder "
			} else if iEncoder.Stats().Setup > 0 && iEncoder.Stats().DataEncoding == 0 {
				s.Prefix = "Encoding data "
			} else if iEncoder.Stats().DataEncoding > 0 && iEncoder.Stats().OutputImageEncoding == 0 {
				s.Prefix = "Generating output PNG image "
			} else {
				break
			}
			time.Sleep(time.Millisecond * 500)
		}
		s.Stop()
	}()

	defer outputFile.Close()
	err = iEncoder.EncodeFiles(filesToHide)
	if err != nil {
		return err
	}
	err = iEncoder.WriteEncodedPNG(outputFile)
	if err != nil {
		return err
	}

	wg.Wait()
	// Change to use logger with global config for log level
	fmt.Printf("Encoder setup time: %s\n", iEncoder.Stats().Setup)
	fmt.Printf("Data encode time: %s\n", iEncoder.Stats().DataEncoding)
	fmt.Printf("Output image encode time: %s\n", iEncoder.Stats().OutputImageEncoding)
	return nil
}

func decodeFilesFromImage() *cobra.Command {
	var encodedImageFile string

	decodeCommand := &cobra.Command{
		Use:     "decode",
		Example: "nsteg image decode --source encoded-image.png",
		Short:   "Decode files from image encoded by nsteg",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.MarkFlagRequired("source"); err != nil {
				return err
			}
			return DecodeFilesFromImage(encodedImageFile)
		},
	}

	decodeCommand.Flags().StringVar(&encodedImageFile, "source", "", "Image generated by nsteg to decode")
	return decodeCommand
}

func DecodeFilesFromImage(encodedMediaFile string) error {
	s := NewSpinner()
	s.Prefix = "Reading source image from disk "
	s.Start()

	srcImage, err := getImageFromFilePath(encodedMediaFile)
	if err != nil {
		return err
	}

	s.Prefix = "Setting up decoder "
	decoder, err := nstegImage.NewImageDecoder(srcImage)
	if err != nil {
		return err
	}

	s.Prefix = "Decoding files "
	decodedFiles, err := decoder.DecodeFiles()
	if err != nil {
		return err
	}

	s.Prefix = "Writing decoded files to disk "
	fileNames := make([]string, 0, len(decodedFiles))
	for _, decodedFile := range decodedFiles {
		fileNames = append(fileNames, decodedFile.Name)
		err = os.WriteFile(decodedFile.Name, decodedFile.Content, 0664)
		if err != nil {
			return err
		}
	}

	s.FinalMSG = fmt.Sprintf("Decoded the following files from the source image: %s\n", strings.Join(fileNames, ","))

	s.Stop()
	return nil
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
