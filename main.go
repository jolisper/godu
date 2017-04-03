package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var size chan int64

type ReqParam struct {
	Dirs string
}

func main() {

	http.HandleFunc("/size", func(rw http.ResponseWriter, r *http.Request) {
		size = make(chan int64)
		defer close(size)

		var p ReqParam
		json.NewDecoder(r.Body).Decode(&p)

		directories := strings.Split(p.Dirs, ",")

		var wg sync.WaitGroup
		wg.Add(len(directories))

		for _, directory := range directories {
			go func(dir string) {
				defer wg.Done()
				walkDir(dir)
			}(directory)
		}

		var total int64

		go func() {
			for s := range size {
				total += s
			}
		}()

		wg.Wait()

		fmt.Fprintf(rw, "%.2f GB\n", float32(total)/1e9)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func walkDir(dirname string) {
	for _, entry := range dirents(dirname) {
		if entry.IsDir() {
			subdir := filepath.Join(dirname, entry.Name())
			walkDir(subdir)
		} else {
			size <- entry.Size()
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
