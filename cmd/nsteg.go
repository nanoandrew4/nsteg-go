package main

import (
	"github.com/spf13/cobra"
	"log"
	"nsteg/internal/cli"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var cpuProfile, memProfileDir string
	rootCommand := &cobra.Command{
		Use:   "nsteg",
		Short: "Steganography application",
	}

	rootCommand.AddCommand(cli.ImageCommands(), cli.ServeAppCommand())

	rootCommand.PersistentFlags().StringVar(&cpuProfile, "cpu-profile", "", "File to which to write the CPU profile")
	rootCommand.PersistentFlags().StringVar(&memProfileDir, "mem-profile-dir", "", "Directory to which to write memory profiles")

	var cpuProfTeardown, memProfTeardown func()
	if cpuProfile != "" {
		cpuProfTeardown = setupCPUProfilingAndReturnTeardown(cpuProfile)
		defer cpuProfTeardown()
	}

	if memProfileDir != "" {
		memProfTeardown = setupMemProfilingAndReturnTeardown(memProfileDir)
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

	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
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
