package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
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
	dlStatus    chan string
	fileName    string
	fileStatus  string
	connections int
	size        int
}

func (dl *download) get(chunkSize int, i int) {
	var start, end int
	req, _ := http.NewRequest("GET", dl.url, nil)
	if i == 0 {
		start = 0
	} else {
		start = (i * chunkSize)
	}

	if i == (dl.connections - 1) {
		end = dl.size
	} else {
		end = start + chunkSize - 1
	}

	fmt.Println("Request from", start, " to ", end)
	req.Header.Set("Range", "bytes="+strconv.Itoa(start)+"-"+strconv.Itoa(end))

	filename := dl.fileName + "." + strconv.Itoa(i)
	out, _ := os.Create(filename)

	client := http.Client{}
	resp, _ := client.Do(req)

	if written, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println("Error occured", err)
		dl.dlStatus <- "failed"
		return
	} else {
		fmt.Println("Success", filename, " written", written)
		dl.dlStatus <- filename + " done"
		return

	}

}
func (dl *download) join() string {
	i := dl.connections - 1
	out, _ := os.Create(dl.fileName)
	defer out.Close()
	for j := 0; j <= i; j++ {

		infile := dl.fileName + "." + strconv.Itoa(j)
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
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	fmt.Println("using %i")
	startTime := time.Now()
	flag.Parse()
	dl := download{
		url:         *urlPtr,
		fileName:    *outFilePtr,
		connections: *nPtr,
		dlStatus:    make(chan string, 1),
	}
	nowTime := time.Now().Format("20060102-150405")
	if dl.fileName == "defaultfilename" {
		dl.fileName = nowTime
	}

	fmt.Println(dl.url)

	resp, _ := http.Get(dl.url)

	dl.size = int(resp.ContentLength)
	isPartial := resp.Header.Get("Accept-Ranges")
	if isPartial == "" {
		fmt.Println("Partial downloads not supported")
		dl.connections = 1
	} else {
		fmt.Println("Partials supported: ", isPartial)
	}
	chunkSize := int(dl.size) / (dl.connections)
	fmt.Println("Content size:", dl.size, "\n Chunk Size ", chunkSize)

	for i := 0; i < dl.connections; i++ {

		go dl.get(chunkSize, i)

	}

	defer resp.Body.Close()

	for j := 0; j < dl.connections; j++ {

		fmt.Println(<-dl.dlStatus)

	}
	timeTaken := time.Since(startTime).Seconds()
	dl.join()

	speed := (float64(dl.size / 1024)) / timeTaken
	fmt.Printf("done time taken:  %.3f seconds , %.2f Speed:kB/s\n", timeTaken, speed)

}
