package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var size chan int64

type ReqParam struct {
	Directories []string `json:"directories"`
}

type Response struct {
	Sizes map[string]string `json:"sizes"`
	Total int64             `json:"total"`
}

func main() {

	http.HandleFunc("/size", func(rw http.ResponseWriter, r *http.Request) {
		size = make(chan int64)
		defer close(size)

		var p ReqParam
		json.NewDecoder(r.Body).Decode(&p)

		var wg sync.WaitGroup
		wg.Add(len(p.Directories))

		for _, directory := range p.Directories {
			go func(dir string) {
				defer wg.Done()
				walkDir(dir, size)
			}(directory)
		}

		var total int64

		sizes := make(map[string]string)
		go func() {
			defer wg.Done()
			for s := range size {
				sizes[strconv.Itoa(rand.Int())] = strconv.Itoa(int(s))
				total += s
			}
		}()
		wg.Wait()

		resp := Response{Sizes: sizes, Total: total}

		fmt.Println(total)
		fmt.Println(float32(total) / 1e9)
		fmt.Println(resp.Total)

		sresp, _ := json.Marshal(resp)
		io.WriteString(rw, string(sresp))
		//string(sresp)

		//		fmt.Fprintf(rw, "%.2f GB\n", float32(total)/1e9)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func walkDir(dirname string, size chan int64) {
	for _, entry := range dirents(dirname) {
		if entry.IsDir() {
			subdir := filepath.Join(dirname, entry.Name())
			walkDir(subdir, size)
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
