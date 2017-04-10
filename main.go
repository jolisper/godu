package main

import (
	"encoding/json"
	"flag"
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

	httpMode := flag.Bool("http-mode", false, "Start a web server, send the query via http: curl -X POST localhost:8080/size -d '{ \"directories\": [\"/my/directory\"] }'")
	flag.Parse()

	if *httpMode {
		httpModeBehaviour()
	} else {
		commandModeBehaviour()
	}
}

func httpModeBehaviour() {
	http.HandleFunc("/size", func(rw http.ResponseWriter, r *http.Request) {
		// Payload decoding
		var p Request
		json.NewDecoder(r.Body).Decode(&p)

		ctotal, err := calculateSize(p.Directories)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

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

func commandModeBehaviour() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "usage: godu directory ...")
		os.Exit(1)
	}

	csize, err := calculateSize(os.Args[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return
	}

	unitOfMeasure := "bytes"

	fmt.Fprintf(os.Stdout, "%d %s\n", <-csize, unitOfMeasure)
}

func calculateSize(directories []string) (chan int64, error) {
	var wg sync.WaitGroup
	wg.Add(len(directories))

	sizeCalculation := func(directory string, csize chan int64) {
		walkDir(directory, csize)
		wg.Done()
	}

	// A Go rutine for every directory argument
	csizes := make(chan int64)
	for _, directory := range directories {
		go sizeCalculation(directory, csizes)
	}

	// Total size calculation
	totalSizeCalculation := func(csizes chan int64, ctotal chan int64) {
		var total int64
		for s := range csizes {
			total += s
		}
		ctotal <- total
		close(ctotal)
	}

	ctotal := make(chan int64)
	go totalSizeCalculation(csizes, ctotal)

	// Wait in other go rutine to not block here
	go func() {
		wg.Wait()
		close(csizes)
	}()

	return ctotal, nil
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
