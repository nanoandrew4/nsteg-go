package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	image "nsteg/image"
	"os"
	"strings"
)

func main() {

	var (
		app = kingpin.New("nsteg", "Steganography application written in Go")

		encode       = app.Command("encode", "Encode files into a media file")
		srcMediaFile = encode.Arg("src-media-file", "Media file to encode data into (original will not be touched)").Required().ExistingFile()
		filesToHide  = encode.Arg("files-to-hide", "Files to hide inside the media file, separated by commas").Required().String()
		outputFile   = encode.Arg("output-file", "Output file path/name for media file with encoded data").Required().String()
		LSBsToUse    = encode.Flag("lsbs-to-use", "Number of least significant bits to use in each channel, in each pixel, for encoding data").Default("5").Int()

		decode           = app.Command("decode", "Decode files from previously encoded media file")
		encodedMediaFile = decode.Arg("encoded-media-file", "Media file containing encoded data").Required().String()
	)

	var err error
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "encode":
		err = image.Encode(*srcMediaFile, *outputFile, strings.Split(*filesToHide, ","), image.Config{LSBsToUse: byte(*LSBsToUse)})
		if err != nil {
			fmt.Printf("Error during image encode: %v", err)
		}
	case "decode":
		image.DecodeImg(*encodedMediaFile)
	}
}
