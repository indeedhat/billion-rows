package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"sync"
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

	inputCh := make(chan []string, 100)
	go readFromFile(inputCh)

	outputCh := make(chan map[string]*DataSet)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go parseData(inputCh, outputCh, &wg)
	}

	dataCh := combineOutputCh(outputCh)

	wg.Wait()
	close(outputCh)

	output(<-dataCh)
}

func parseData(inputCh chan []string, outputCh chan map[string]*DataSet, wg *sync.WaitGroup) {
	defer wg.Done()

	results := make(map[string]*DataSet)
	parts := make([]string, 2)

	for batch := range inputCh {
		for _, line := range batch {
			doSplitOnSemi(line, parts)

			entry, ok := results[string(parts[0])]
			if !ok {
				entry = &DataSet{
					Min: math.MaxFloat64,
					Max: math.SmallestNonzeroFloat64,
				}
				results[string(parts[0])] = entry
			}

			f, _ := strconv.ParseFloat(parts[1], 64)

			entry.Count++
			entry.Total += f
			if f < entry.Min {
				entry.Min = f
			}
			if f > entry.Max {
				entry.Max = f
			}
		}
	}

	outputCh <- results
}

func doSplitOnSemi(str string, out []string) {
	for i := 0; i < 1000; i++ {
		if str[i] == ';' {
			out[0] = str[:i]
			out[1] = str[i+1:]
			return
		}
	}
}

func readFromFile(inputCh chan []string) {
	fh, err := os.Open("dataset/measurements.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	defer close(inputCh)

	var (
		i       int
		lines   = make([]string, 0, 100)
		scanner = bufio.NewScanner(fh)
	)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		i++
		if i == 100 {
			inputCh <- lines
			lines = make([]string, 0, 100)
			i = 0
		}
	}

	if len(lines) > 0 {
		inputCh <- lines
	}
}

func combineOutputCh(outputCh chan map[string]*DataSet) chan map[string]*DataSet {
	resultCh := make(chan map[string]*DataSet, 1)

	go func() {
		final := make(map[string]*DataSet)

		for result := range outputCh {
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
				} else {
					final[k] = v
				}
			}
		}

		resultCh <- final
	}()

	return resultCh
}

func output(data map[string]*DataSet) {
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
