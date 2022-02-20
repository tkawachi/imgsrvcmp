package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/h2non/bimg"
)

// This program compares the responses from the two image servers.
// The base URL of the image server is given as an argument.
// The list of paths to access is given as a file.
// For each path, it records the response of each image server.

type GetResult struct {
	ElaspedMillis   int64               `json:"elasped_millis"`
	StatusCode      int                 `json:"status_code"`
	ResponseHeaders map[string][]string `json:"response_headers"`
	ImageType       string              `json:"image_type"`
	Height          int                 `json:"height"`
	Width           int                 `json:"width"`
}

type CaseResult struct {
	CaseNo  int        `json:"case_no"`
	Result1 *GetResult `json:"result1"`
	Result2 *GetResult `json:"result2"`
}

// Read paths from a file.
func readPaths(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var paths []string
	for {
		var path string
		_, err := fmt.Fscanln(file, &path)
		if err != nil {
			break
		}
		paths = append(paths, path)
	}
	return paths
}

// issues a GET rquest to the given URL and saves the response body.
func httpGet(url string, outFile string) (*GetResult, error) {

	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	elapsed := time.Since(start)

	file, err := os.Create(outFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes := buf.Bytes()
	_, err = file.Write(bytes)
	if err != nil {
		return nil, err
	}
	img := bimg.NewImage(bytes)
	size, err := img.Size()
	if err != nil {
		size = bimg.ImageSize{Height: -1, Width: -1}
	}

	result := &GetResult{
		ElaspedMillis:   int64(elapsed) / int64(time.Millisecond),
		StatusCode:      resp.StatusCode,
		ResponseHeaders: resp.Header,
		ImageType:       bimg.DetermineImageTypeName(bytes),
		Height:          size.Height,
		Width:           size.Width,
	}

	return result, nil
}

// Generates output file path for a given path.
func genOutFileName(caseNo int, servNo int) string {
	return fmt.Sprintf("%d_%d", caseNo, servNo)
}

func main() {
	if len(os.Args) != 4 {
		panic("Usage: imgsrvcmp <baseURL1> <baseURL2> <path_list.yaml>")
	}
	baseURL1 := os.Args[1]
	baseURL2 := os.Args[2]
	paths := readPaths(os.Args[3])

	for i, path := range paths {
		log.Println("Case", i, ":", path)
		outFile1 := genOutFileName(i, 1)
		outFile2 := genOutFileName(i, 2)
		result1, err := httpGet(baseURL1+path, outFile1)
		if err != nil {
			panic(err)
		}
		result2, err := httpGet(baseURL2+path, outFile2)
		if err != nil {
			panic(err)
		}
		caseResult := &CaseResult{
			CaseNo:  i,
			Result1: result1,
			Result2: result2,
		}
		// save caseResult to file as JSON
		file, err := os.Create(fmt.Sprintf("%d.txt", i))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		bytes, err := json.Marshal(caseResult)
		if err != nil {
			panic(err)
		}
		_, err = file.Write(bytes)
		if err != nil {
			panic(err)
		}
	}
}
