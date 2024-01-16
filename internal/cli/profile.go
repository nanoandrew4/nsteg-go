package cli

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	// MemorySampleRate How often to dump the memory to a file in HZ. Values of less than 1 are recommended to avoid
	// having to sort through too many dump files
	MemorySampleRate = 0.5

	cpuProfiler *CPUProfilerStruct
	memProfiler *MemProfilerStruct
)

type CPUProfilerStruct struct {
	profileOutput io.Writer
}

type MemProfilerStruct struct {
	dumpPath           string
	heapDumps          [][]byte
	shouldProfilerStop chan bool
}

func StartCPUProfiler(profileOutput io.Writer) {
	cpuProfiler = &CPUProfilerStruct{profileOutput: profileOutput}
	runtime.SetCPUProfileRate(500)
	err := pprof.StartCPUProfile(profileOutput)
	if err != nil {
		log.Fatalln("Error starting CPU profiler")
	}
}

func StopCPUProfiler() {
	pprof.StopCPUProfile()
	cpuProfiler = nil
}

func StartMemoryProfiler(profileDumpPath string) {
	if MemorySampleRate <= 0 {
		return
	}

	memProfiler = &MemProfilerStruct{dumpPath: profileDumpPath, shouldProfilerStop: make(chan bool)}

	go func() {
		ticker := time.NewTicker(time.Duration((1/MemorySampleRate)*1000) * time.Millisecond)
		for {
			select {
			case <-memProfiler.shouldProfilerStop:
				return
			case <-ticker.C:
				DumpMemoryProfile()
			}
		}
	}()
}

func DumpMemoryProfile() {
	if memProfiler != nil {
		w := bytes.NewBuffer(nil)
		pprof.WriteHeapProfile(w)
		memProfiler.heapDumps = append(memProfiler.heapDumps, w.Bytes())
	}
}

func StopMemoryProfiler() {
	if memProfiler != nil {
		memProfiler.shouldProfilerStop <- true
		DumpMemoryProfile()
		_ = os.Mkdir(memProfiler.dumpPath, os.ModePerm)
		for dIdx, dump := range memProfiler.heapDumps {
			err := os.WriteFile(fmt.Sprintf("%s/mem-%d.mprof", memProfiler.dumpPath, dIdx), dump, os.ModePerm)
			if err != nil {
				log.Println("Error writing memory profile to disk")
			}
		}
	}
}
