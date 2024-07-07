package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
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

	data := load()
	output(data)
}

func load() map[string]*DataSet {
	resuts := make(map[string]*DataSet)

	fh, err := os.Open("dataset/measurements.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ";")

		if len(parts) != 2 {
			log.Fatalf("found bad line, %s", line)
		}

		entry, ok := resuts[string(parts[0])]
		if !ok {
			entry = &DataSet{
				Min: math.MaxFloat64,
				Max: math.SmallestNonzeroFloat64,
			}
			resuts[string(parts[0])] = entry
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

	return resuts
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
