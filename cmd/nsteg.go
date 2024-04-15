package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"image/png"
	"log"
	"nsteg/internal/cli"
	"nsteg/internal/server"
	"nsteg/pkg/config"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	var (
		app           = kingpin.New("nsteg", "Steganography application written in Go")
		cpuProfile    = app.Flag("cpu-profile", "Dump CPU profile into the supplied file").String()
		memProfileDir = app.Flag("mem-profile-dir", "Dump memory profiles into the supplied directory").String()

		encode              = app.Command("encode", "Encode files into a media file")
		srcMediaFile        = encode.Flag("src-media-file", "Media file to encode data into (original will not be touched)").Required().ExistingFile()
		filesToHide         = encode.Flag("files-to-hide", "Files to hide inside the media file, separated by commas").Required().String()
		outputFile          = encode.Flag("output-file", "Output file path/name for media file with encoded data").Required().String()
		LSBsToUse           = encode.Flag("lsbs-to-use", "Number of least significant bits to use in each channel, in each pixel, for encoding data").Default("3").Int()
		imageEnc            = encode.Command("image", "Encode image file opts")
		pngCompressionLevel = imageEnc.Flag("png-compression-level", "Level of compression for png, 0 is default, -1 is no compression, -2 is fast compression, -3 is best compression").Default("-3").Int()

		decode           = app.Command("decode", "Decode files from previously encoded media file")
		encodedMediaFile = decode.Arg("encoded-media-file", "Media file containing encoded data").Required().String()

		serve = app.Command("serve", "Start webserver")
		port  = serve.Flag("port", "Port to start server on").Default("8080").String()
	)

	parsedArgs := kingpin.MustParse(app.Parse(os.Args[1:]))

	var cpuProfTeardown, memProfTeardown func()
	if cpuProfile != nil && *cpuProfile != "" {
		cpuProfTeardown = setupCPUProfilingAndReturnTeardown(*cpuProfile)
		defer cpuProfTeardown()
	}

	if memProfileDir != nil && *memProfileDir != "" {
		memProfTeardown = setupMemProfilingAndReturnTeardown(*memProfileDir)
		defer memProfTeardown()
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // subscribe to system signals
	onKill := func(c chan os.Signal) {
		select {
		case <-c:
			if cpuProfTeardown != nil {
				cpuProfTeardown()
			}
			if memProfTeardown != nil {
				memProfTeardown()
			}
			os.Exit(0)
		}
	}

	go onKill(c)

	var err error
	switch parsedArgs {
	case "encode image":
		err = cli.EncodeImageWithFiles(*srcMediaFile, *outputFile, strings.Split(*filesToHide, ","), config.ImageEncodeConfig{
			LSBsToUse:           byte(*LSBsToUse),
			PngCompressionLevel: png.CompressionLevel(*pngCompressionLevel),
		})
		if err != nil {
			fmt.Printf("Error during image encode: %v\n", err)
		}
	case "decode":
		err = cli.DecodeFilesFromImage(*encodedMediaFile)
		if err != nil {
			fmt.Printf("Error during image decode: %v\n", err)
		}
	case "serve":
		server.StartServer(*port)
	default:
		log.Fatal("Unknown options")
	}
}

func setupCPUProfilingAndReturnTeardown(cpuProfile string) (deferredTeardown func()) {
	cpuProfileFile, err := os.Create(cpuProfile)
	if err != nil {
		log.Fatal(err)
	}
	cli.StartCPUProfiler(cpuProfileFile)

	return func() {
		cli.StopCPUProfiler()
		cpuProfileFile.Close()
	}
}

func setupMemProfilingAndReturnTeardown(memProfileDir string) (deferredTeardown func()) {
	cli.StartMemoryProfiler(memProfileDir)
	return func() {
		cli.StopMemoryProfiler()
	}
}
