package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"image/png"
	"log"
	image "nsteg/image"
	"os"
	"strings"
)

func main() {

	var (
		app = kingpin.New("nsteg", "Steganography application written in Go")

		encode              = app.Command("encode", "Encode files into a media file")
		srcMediaFile        = encode.Flag("src-media-file", "Media file to encode data into (original will not be touched)").Required().ExistingFile()
		filesToHide         = encode.Flag("files-to-hide", "Files to hide inside the media file, separated by commas").Required().String()
		outputFile          = encode.Flag("output-file", "Output file path/name for media file with encoded data").Required().String()
		LSBsToUse           = encode.Flag("lsbs-to-use", "Number of least significant bits to use in each channel, in each pixel, for encoding data").Default("3").Int()
		imageEnc            = encode.Command("image", "Encode image file opts")
		pngCompressionLevel = imageEnc.Flag("png-compression-level", "Level of compression for png, 0 is default, -1 is no compression, -2 is fast compression, -3 is best compression").Default("-3").Int()

		decode           = app.Command("decode", "Decode files from previously encoded media file")
		encodedMediaFile = decode.Arg("encoded-media-file", "Media file containing encoded data").Required().String()
	)

	var err error
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "encode image":
		err = image.Encode(*srcMediaFile, *outputFile, strings.Split(*filesToHide, ","), image.Config{
			LSBsToUse:           byte(*LSBsToUse),
			PngCompressionLevel: png.CompressionLevel(*pngCompressionLevel),
		})
		if err != nil {
			fmt.Printf("Error during image encode: %v\n", err)
		}
	case "decode":
		err = image.DecodeImg(*encodedMediaFile)
		if err != nil {
			fmt.Printf("Error during image decode: %v\n", err)
		}
	default:
		log.Fatal("Unknown options")
	}
}
