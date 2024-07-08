package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"strconv"
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
		chunkChan   = make(chan []byte, 10)
		resultsChan = make(chan map[string]StationData, 10)
		finalChan   = make(chan map[string]StationData, 1)
	)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go parseChunkData(chunkChan, resultsChan, &wg)
	}

	go readChunkFromFile(chunkChan)
	go func() {
		results := make(map[string]StationData, 1000)
		for r := range resultsChan {
			for stationName, e := range r {
				if entry, ok := results[stationName]; ok {
					entry.Count += e.Count
					entry.Total += e.Total
					entry.Max = math.Max(entry.Max, e.Max)
					entry.Min = math.Min(entry.Min, e.Min)
					results[stationName] = entry
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

func formatOutput(stationData map[string]StationData) {
	var (
		buf  bytes.Buffer
		keys []string
	)

	for k := range stationData {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	buf.WriteByte('{')

	for i, k := range keys {
		entry := stationData[k]
		buf.WriteString(fmt.Sprintf("%s=%.1f,%.1f,%.1f", k, entry.Min, entry.Total/float64(entry.Count), entry.Max))

		if i < len(keys)-1 {
			buf.WriteString(", ")
		}
	}

	buf.WriteByte('}')

	fmt.Print(buf.String())
}

func readChunkFromFile(chunkChan chan []byte) {
	fh, err := os.OpenFile(FileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal("failed to open dataset: ", err)
	}
	defer fh.Close()

	var (
		extraData []byte
	)
	for {
		chunkData := make([]byte, ChunkSize)

		readLen, err := fh.Read(chunkData)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}

		lastNewline := bytes.LastIndex(chunkData[:readLen], []byte{'\n'})

		chunkChan <- append(extraData, chunkData[:lastNewline]...)

		extraData = make([]byte, readLen-lastNewline)
		copy(extraData, chunkData[lastNewline+1:])
	}

	close(chunkChan)
}

type StationData struct {
	Min   float64
	Max   float64
	Total float64
	Count int
}

func parseChunkData(chunkChan chan []byte, resultsChan chan map[string]StationData, wg *sync.WaitGroup) {
	defer wg.Done()

	for chunkData := range chunkChan {
		var (
			cursor      int
			stationName string
			stringData  = string(chunkData)
			results     = make(map[string]StationData)
		)
		for i, char := range stringData {
			if char == ';' {
				stationName = stringData[cursor:i]
				cursor = i + 1
			} else if char == '\n' {
				temp, _ := strconv.ParseFloat(stringData[cursor:i], 64)
				cursor = i + 2

				if entry, ok := results[stationName]; ok {
					entry.Count++
					entry.Total += temp
					entry.Max = math.Max(entry.Max, temp)
					entry.Min = math.Min(entry.Min, temp)
					results[stationName] = entry
					continue
				}

				results[stationName] = StationData{
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
