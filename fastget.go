package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/cheggaaa/pb"
)

var (
	nPtr       = flag.Int("n", 4, " Number of connections")
	urlPtr     = flag.String("url", "", "Url")
	outFilePtr = flag.String("O", "defaultfilename", "Output file name")
)

type download struct {
	url         string
	dlStatus    chan string
	fileName    string
	fileStatus  string
	connections int
	size        int
	client      *http.Client

	pb           *pb.ProgressBar
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
	reqRange := "bytes=" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
	//fmt.Println("Request :", reqRange)
	req.Header.Set("Range", reqRange)



	filename := dl.fileName + "." + strconv.Itoa(i)
	out, _ := os.Create(filename)

	client := dl.client
	resp, _ := client.Do(req)

	writer := io.MultiWriter(out, dl.pb)


	_, err := io.Copy(writer,resp.Body)

	if err != nil {
		fmt.Println("Error occured", err)
		dl.dlStatus <- "failed"
		return
	} else {
		//fmt.Println("Success", filename, " written", written)
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
		writer := io.MultiWriter(out, dl.pb)


		_, err := io.Copy(writer,in)


		if err != nil {
			fmt.Println("Error occured", err)

		} else {
			//fmt.Println("file ", infile, "joined")

		}

	}

	fmt.Println("file join complete")
	return "done"
}

func runTime(timeElapsed float64) string {
	var pTime string

	if timeElapsed > 60 {
		mins := int(timeElapsed / 60)
		seconds := int(math.Mod(timeElapsed, 60))
		pTime = strconv.Itoa(mins) + " mins " + strconv.Itoa(seconds) + " seconds"
	} else {
		pTime = strconv.FormatFloat(timeElapsed, 'f', 2, 64) + " seconds"
	}
	return pTime
}
func main() {
	cpuCount := runtime.NumCPU()
	i := runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Was using ", i, " CPUs changing to ", cpuCount)
	startTime := time.Now()
	flag.Parse()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}


	dl := download{
		url:         *urlPtr,
		fileName:    *outFilePtr,
		connections: *nPtr,
		dlStatus:    make(chan string, 1),
		client:  &http.Client{Transport: tr},
	}
	//nowTime := time.Now().Format("20060102-150405")
	if dl.url == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter a valid url: ")
		dl.url, _ = reader.ReadString('\n')
		dl.url=strings.Replace(strings.Replace(dl.url,"\n","",-1),"\r","",-1)
		fmt.Println(dl.url)
	}
	if dl.fileName == "defaultfilename" {
		//dl.fileName = nowTime
		i := strings.LastIndex(dl.url, "/")
		dl.fileName = dl.url[i+1:]

	}


	fmt.Println(dl.url)

	resp, err := dl.client.Get(dl.url)

	if err != nil {
		fmt.Println(err)
	}
	dl.size = int(resp.ContentLength)
	isPartial := resp.Header.Get("Accept-Ranges")
	if isPartial == "" {
		fmt.Println("Partial downloads not supported",resp.Body)
		dl.connections = 1
	} else {
		fmt.Println("Partials supported: ", isPartial)
	}
	chunkSize := int(dl.size) / (dl.connections)
	dl.pb = pb.New(dl.size).SetUnits(pb.U_BYTES)
	dl.pb.Start()
	fmt.Println("Content size:", dl.size, "\n Chunk Size ", chunkSize)

	for i := 0; i < dl.connections; i++ {

		go dl.get(chunkSize, i)

	}

	defer resp.Body.Close()

	for j := 0; j < dl.connections; j++ {

		<-dl.dlStatus

	}
	close(dl.dlStatus)
	timeTaken := time.Since(startTime).Seconds()
	timeTakenS := runTime(timeTaken)
	dl.pb = pb.New(dl.size).SetUnits(pb.U_BYTES)
	dl.pb.Start()
	dl.join()

	speed := (float64(dl.size / 1024)) / timeTaken
	fmt.Printf("done time taken:  %s , %.4f Speed:kB/s\n", timeTakenS, speed)

}
