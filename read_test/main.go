package main

import (
	"log"
	"os"
)

func main() {
	fh, err := os.Open("read_test/example.text")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	chunk := make([]byte, 500)

	for {
		readLen, err := fh.Read(chunk)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("readLen(%d) len(%d) cap(%d)", readLen, len(chunk), cap(chunk))

	}
}
