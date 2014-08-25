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

var nPtr = flag.Int("n", 4, " Number of connections")
var urlPtr = flag.String("url", "http://wprfu.net/~shenal/largefile", "Url")

type download struct {
	url         string
	dlStatus      string
	fileName    string
	fileStatus string
	connections int
}


func get(req *http.Request, filename string, status chan string) {
	out, _ := os.Create(filename)

	client := &http.Client{}
	resp, _ := client.Do(req)
	if written, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println("Error occured", err)
		status <- "failed"
	} else {
		fmt.Println("Success", filename, " written", written)
		status <- filename + " done"

	}

}
func join(i int, filename string) string {
	out, _ := os.Create(filename)

	for j := 0; j <= i; j++ {

		infile := filename + "." + strconv.Itoa(j)
		in, _ := os.Open(infile)

		buffer, err := io.Copy(out, in)
		//defer os.Remove(infile)

		buffer = 0
		if err != nil {
			fmt.Println("Error occured", err)

		} else {
			fmt.Println("file ", j, " written", buffer)
			in = nil
		}
	}
	out.Close()
	fmt.Println("file join complete")
	return "done"
}
func main() {
	nowTime := time.Now().Format("20060102-150405")
	flag.Parse()
	url := *urlPtr
	fmt.Println(url)

	n := *nPtr - 1
	fmt.Println(n)
	status := make(chan string, 1)
	resp, _ := http.Get(*urlPtr)
	i := 0
	if isPartial := resp.Header.Get("Accept-Ranges"); isPartial != "" {
		fileSize := resp.ContentLength
		chunkSize := int(fileSize) / (n + 1)
		fmt.Println("Partials supportedL: ", isPartial, "\n Content size:", fileSize, "\n Chunk Size ", chunkSize)
		end := chunkSize
		start := 0

		for i = 0; i < n; i++ {

			end = start + chunkSize
			fmt.Println("Request from", start, " to ", end)
			req, _ := http.NewRequest("GET", *urlPtr, nil)
			req.Header.Set("Range", "bytes="+strconv.Itoa(start)+"-"+strconv.Itoa(end))
			go get(req, nowTime+"."+strconv.Itoa(i), status)
			start += chunkSize + 1

		}
		startOfEnd := end + 1
		fmt.Println("Request from", end+1, " to ", fileSize)
		req, _ := http.NewRequest("GET", *urlPtr, nil)
		req.Header.Set("Range", "bytes="+strconv.Itoa(startOfEnd)+"-"+strconv.Itoa(int(fileSize)))
		go get(req, nowTime+"."+strconv.Itoa(n), status)
	} else {
		fmt.Println("error occured")

	}
	defer resp.Body.Close()
	for j := 0; j <= i; j++ {

		fmt.Println(<-status)

	}
	join(i, nowTime)
	fmt.Println("done done done")

}
