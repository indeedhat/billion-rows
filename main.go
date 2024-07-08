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
	"sort"
	"strconv"
	"sync"
)

const (
	ChunkSize   = 32 * 1024 * 1024
	DatasetPath = "dataset/measurements.txt"
)

type DataSet struct {
	Min   float64
	Max   float64
	Total float64
	Count int
}

var (
	profileCpu    bool
	profileMemory bool
)

func main() {
	flag.BoolVar(&profileCpu, "profile-cpu", false, "Create a cpu profile for the run")
	flag.BoolVar(&profileMemory, "profile-mem", false, "Create a memory profile for the run")
	flag.Parse()

	if profileCpu {
		log.Println("starting profiler")
		cpuFh, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal("Failed to create cpu profile: ", err)
		}
		defer cpuFh.Close()

		if err := pprof.StartCPUProfile(cpuFh); err != nil {
			log.Fatal("failed to start cpu profiler: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	var (
		readWg      sync.WaitGroup
		splitWg     sync.WaitGroup
		readChan    = make(chan []byte, 10)
		linesChan   = make(chan []string, 100)
		resultsChan = make(chan map[string]DataSet, 100)
	)

	for i := 0; i < (runtime.NumCPU()/2)-1; i++ {
		readWg.Add(1)
		go readLines(i, linesChan, resultsChan, &readWg)
	}

	for i := 0; i < (runtime.NumCPU()/2)-1; i++ {
		splitWg.Add(1)
		go splitChunks(readChan, linesChan, &splitWg)
	}

	go readChunks(readChan)

	log.Println("waiting for split")
	splitWg.Wait()

	close(linesChan)

	log.Println("waiting for read")
	readWg.Wait()

	log.Println("final")
	finalChan := combineOutputCh(resultsChan)

	close(resultsChan)

	output(<-finalChan)
}

func doMemProfile(i int) {
	log.Println("starting memory")
	memFh, err := os.Create(fmt.Sprintf("mem_%d.prof", i))
	if err != nil {
		log.Fatal("Failed to create mem profile: ", err)
	}
	defer memFh.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(memFh); err != nil {
		log.Fatal("failed to run memory profiler: ", err)
	}
}

func readLines(
	idx int,
	linesChan chan []string,
	resultsChan chan map[string]DataSet,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	results := make(map[string]DataSet)

	for batch := range linesChan {
		for _, line := range batch {
			for i := 0; i < len(line); i++ {
				if line[i] != ';' {
					continue
				}

				key := line[:i]

				f, _ := strconv.ParseFloat(line[i+1:], 64)
				if entry, ok := results[key]; ok {
					entry.Count++
					entry.Total += f

					if f < entry.Min {
						entry.Min = f
					}
					if f > entry.Max {
						entry.Max = f
					}

					results[key] = entry
					continue
				}

				if len(results) > 420 {
					continue
				}
				results[key] = DataSet{
					Min:   f,
					Max:   f,
					Total: f,
					Count: 1,
				}

			}
		}
	}

	resultsChan <- results
}

func splitChunks(readChan chan []byte, linesChan chan []string, wg *sync.WaitGroup) {
	defer wg.Done()
	for chunk := range readChan {
		var (
			lastIndex int
			buf       = string(chunk)
			lines     = make([]string, 0, 100)
		)

		for i := range chunk {
			if chunk[i] != '\n' {
				continue
			}

			lines = append(lines, buf[lastIndex+1:i])
			lastIndex = i

			if len(lines) == 100 {
				linesChan <- lines
				lines = make([]string, 0, 100)
			}
		}

		if len(lines) > 0 {
			linesChan <- lines
		}
	}
}

func readChunks(readChan chan []byte) {
	fh, err := os.Open(DatasetPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	defer close(readChan)

	var (
		extraData = []byte{}
		chunk     = make([]byte, ChunkSize)
	)
	for {
		readLen, err := fh.Read(chunk)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatal(err)
			}
		}

		newlineIdx := bytes.LastIndex(chunk[:readLen], []byte{'\n'})

		buffer := make([]byte, newlineIdx)
		copy(buffer, chunk)
		buffer = append(extraData, chunk[:newlineIdx+1]...)

		readChan <- buffer

		extraData = make([]byte, readLen-(newlineIdx+1))
		copy(extraData, chunk[newlineIdx+1:readLen])
	}
}

func combineOutputCh(resultsChan chan map[string]DataSet) chan map[string]DataSet {
	finalChan := make(chan map[string]DataSet, 1)

	go func() {
		final := make(map[string]DataSet)

		for result := range resultsChan {
			for k, v := range result {
				if entry, ok := final[k]; ok {
					entry.Count += v.Count
					entry.Total += v.Total

					if v.Min < entry.Min {
						entry.Min = v.Min
					}
					if v.Max > entry.Max {
						entry.Max = v.Max
					}
					final[k] = entry
				} else {
					final[k] = v
				}
			}
		}

		finalChan <- final
	}()

	return finalChan
}

func output(data map[string]DataSet) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, key := range keys {
		buf.WriteString(fmt.Sprintf("%s=%.1f,%.1f,%.1f",
			key,
			data[key].Min,
			data[key].Total/float64(data[key].Count),
			data[key].Max,
		))

		if i < len(keys)-1 {
			buf.WriteString(", ")
		}
	}

	buf.WriteByte('}')
	fmt.Print(buf.String())
}
