package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"sync"
)

const (
	FileName  = "dataset/measurements.txt"
	ChunkSize = 32 * 1024 * 1024
)

var cpuProfile bool

func main() {
	flag.BoolVar(&cpuProfile, "profile-cpu", false, "generate a cpu profile")
	flag.Parse()

	if cpuProfile {
		fh, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal("failed to create pprof file: ", err)
		}

		if err := pprof.StartCPUProfile(fh); err != nil {
			log.Fatal("failed to start cpu profiler: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	var (
		wg          sync.WaitGroup
		chunkChan   = make(chan string, 10)
		resultsChan = make(chan map[string]*StationData, 10)
		finalChan   = make(chan map[string]*StationData, 1)
	)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go parseChunkData(chunkChan, resultsChan, &wg)
	}

	go readChunkFromFile(chunkChan)
	go func() {
		results := make(map[string]*StationData, 1000)
		for r := range resultsChan {
			for stationName, e := range r {
				if entry, ok := results[stationName]; ok {
					entry.Count += e.Count
					entry.Total += e.Total
					if e.Max > entry.Max {
						entry.Max = e.Max
					}
					if e.Min < entry.Min {
						entry.Min = e.Min
					}
				} else {
					results[stationName] = e
				}
			}
		}

		finalChan <- results
	}()

	wg.Wait()
	close(resultsChan)

	formatOutput(<-finalChan)
}

func formatOutput(stationData map[string]*StationData) {
	var (
		buf  bytes.Buffer
		keys []string
	)

	for k := range stationData {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for i, k := range keys {
		entry := stationData[k]
		buf.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f",
			k, float64(entry.Min)/10,
			(float64(entry.Total)/10)/float64(entry.Count),
			float64(entry.Max)/10),
		)

		if i < len(keys)-1 {
			buf.WriteString(", ")
		}
	}

	fmt.Print(buf.String())
}

func readChunkFromFile(chunkChan chan string) {
	fh, err := os.OpenFile(FileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("failed to open dataset: ", err)
	}
	defer fh.Close()

	var (
		extraData []byte
		chunkData = make([]byte, ChunkSize)
	)
	for {
		readLen, err := fh.Read(chunkData)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}

		lastNewline := bytes.LastIndex(chunkData[:readLen], []byte{'\n'})

		chunkChan <- string(append(extraData, chunkData[:lastNewline]...))

		extraData = make([]byte, readLen-lastNewline-1)
		copy(extraData, chunkData[lastNewline+1:readLen])
	}

	close(chunkChan)
}

type StationData struct {
	Min   int64
	Max   int64
	Total int64
	Count int
}

func parseChunkData(chunkChan chan string, resultsChan chan map[string]*StationData, wg *sync.WaitGroup) {
	defer wg.Done()

	for stringData := range chunkChan {
		var (
			cursor      int
			stationName string
			results     = make(map[string]*StationData)
		)

		for i, char := range stringData {
			if char == ';' {
				stationName = stringData[cursor:i]
				cursor = i + 1
			} else if char == '\n' {
				temp := parseTemp(stringData[cursor:i])
				cursor = i + 1

				if entry, ok := results[stationName]; ok {
					entry.Count++
					entry.Total += temp
					if temp > entry.Max {
						entry.Max = temp
					}
					if temp < entry.Min {
						entry.Min = temp
					}
					continue
				}

				results[stationName] = &StationData{
					Min:   temp,
					Max:   temp,
					Total: temp,
					Count: 1,
				}
			}
		}

		stringData = ""
		resultsChan <- results
	}
}

func parseTemp(temp string) int64 {
	var (
		offset   int
		negative bool
	)

	if temp[0] == '-' {
		offset = 1
		negative = true
	}

	w := parseUint(temp[offset:])

	if negative {
		return -int64(w)
	}

	return int64(w)
}

// parseUint64 is a stripped down version of strconv.ParseUint
// i removed everythin not necessary for this task
func parseUint(s string) uint64 {
	base := 10
	var n uint64
	for _, c := range []byte(s) {
		if c == '.' {
			continue
		}
		d := c - '0'
		n *= uint64(base)
		n += uint64(d)
	}

	return n
}
