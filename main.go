package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type Request struct {
	Directories []string `json:"directories"`
}

type Response struct {
	Directories   []string `json:"directories"`
	TotalSize     float64  `json:"totalSize"`
	UnitOfMeasure string   `json:"unitOfMeasure"`
}

func main() {

	http.HandleFunc("/size", func(rw http.ResponseWriter, r *http.Request) {
		// Payload decoding
		var p Request
		json.NewDecoder(r.Body).Decode(&p)

		// A Go rutine for every directory
		var wg sync.WaitGroup
		wg.Add(len(p.Directories))

		csize := make(chan int64)
		for _, directory := range p.Directories {
			go func(dir string) {
				defer wg.Done()
				walkDir(dir, csize)
			}(directory)
		}

		// Total calculation
		ctotal := make(chan int64)
		go func() {
			var total int64
			for s := range csize {
				total += s
			}
			ctotal <- total
			close(ctotal)
		}()

		wg.Wait()
		close(csize)

		// Response payload building
		response := Response{
			Directories:   p.Directories,
			TotalSize:     float64(<-ctotal),
			UnitOfMeasure: "bytes",
		}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		// Response sending
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		rw.Write(js)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func walkDir(dirname string, csize chan int64) {
	for _, entry := range dirents(dirname) {
		if entry.IsDir() {
			subdir := filepath.Join(dirname, entry.Name())
			walkDir(subdir, csize)
		} else {
			csize <- entry.Size()
		}
	}
}

func dirents(dirname string) []os.FileInfo {
	dirFiles, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil
	}
	return dirFiles
}
