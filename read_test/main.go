package main

import (
	"log"
	"os"
)

func main() {
	fh, err := os.Open("dataset/measurements.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	chunk := make([]byte, 1024*1024*128)

	for {
		len, err := fh.Read(chunk)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(len)
	}
}
