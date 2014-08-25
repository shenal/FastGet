package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	nPtr       = flag.Int("n", 4, " Number of connections")
	urlPtr     = flag.String("url", "http://wprfu.net/~shenal/largefile", "Url")
	outFilePtr = flag.String("O", "defaultfilename", "Output file name")
)

type download struct {
	url         string
	dlStatus    string
	fileName    string
	fileStatus  string
	connections int
}

func get(req *http.Request, filename string, status chan string) {
	out, _ := os.Create(filename)

	client := &http.Client{}
	resp, _ := client.Do(req)

	if written, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println("Error occured", err)
		status <- "failed"
		return
	} else {
		fmt.Println("Success", filename, " written", written)
		status <- filename + " done"
		return

	}

}
func join(i int, filename string) string {
	out, _ := os.Create(filename)
	defer out.Close()
	for j := 0; j <= i; j++ {

		infile := filename + "." + strconv.Itoa(j)
		in, _ := os.Open(infile)
		defer os.Remove(infile)

		_, err := io.Copy(out, in)

		if err != nil {
			fmt.Println("Error occured", err)

		} else {
			fmt.Println("file ", infile, "joined")

		}

	}

	fmt.Println("file join complete")
	return "done"
}
func main() {
	nowTime := time.Now().Format("20060102-150405")
	outFileName := *outFilePtr
	if outFileName == "defaultfilename" {
		outFileName = nowTime
	}
	flag.Parse()
	url := *urlPtr
	fmt.Println(url)
	startTime := time.Now()

	n := *nPtr - 1
	fmt.Println(n)
	status := make(chan string, 1)
	resp, _ := http.Get(*urlPtr)
	i := 0
	fileSize := resp.ContentLength
	if isPartial := resp.Header.Get("Accept-Ranges"); isPartial != "" {

		chunkSize := int(fileSize) / (n + 1)
		fmt.Println("Partials supported: ", isPartial, "\n Content size:", fileSize, "\n Chunk Size ", chunkSize)
		end := chunkSize
		start := 0

		for i = 0; i < n; i++ {

			end = start + chunkSize
			fmt.Println("Request from", start, " to ", end)
			req, _ := http.NewRequest("GET", *urlPtr, nil)
			req.Header.Set("Range", "bytes="+strconv.Itoa(start)+"-"+strconv.Itoa(end))
			go get(req, outFileName+"."+strconv.Itoa(i), status)
			start += chunkSize + 1

		}
		startOfEnd := end + 1
		fmt.Println("Request from", end+1, " to ", fileSize)
		req, _ := http.NewRequest("GET", *urlPtr, nil)
		req.Header.Set("Range", "bytes="+strconv.Itoa(startOfEnd)+"-"+strconv.Itoa(int(fileSize)))
		go get(req, nowTime+"."+strconv.Itoa(n), status)
	} else {
		fmt.Println("Partial downloads not supported")

	}
	defer resp.Body.Close()

	for j := 0; j <= i; j++ {

		fmt.Println(<-status)

	}
	join(i, nowTime)
	timeTaken := time.Since(startTime).Seconds()
	speed := (float64(fileSize / 1024)) / timeTaken
	fmt.Printf("done time taken:  %.3f seconds , %.2f Speed:kB/s", timeTaken, speed)

}
