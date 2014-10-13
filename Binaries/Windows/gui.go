package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/cheggaaa/pb"
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
)

func main() {
	var url, connCount, outTE *walk.TextEdit
	MainWindow{
		Title:   "FastGet",
		MinSize: Size{550, 200},
		Layout:  VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					VSplitter{
						MaxSize: Size{50, 200},
						Children: []Widget{
							Label{Text: "URL"},
							Label{},
							Label{},
							Label{},
							Label{Text: "Number of connections"},
							Label{},
							Label{},
							Label{},
							Label{},
							Label{},
							Label{},
							Label{},
						},
					},
					VSplitter{
						MinSize: Size{250, 200},
						Children: []Widget{
							TextEdit{AssignTo: &url, Text: "http://google.com"},
							TextEdit{AssignTo: &connCount, Text: "4"},
							PushButton{
								Text: "GET!!",
								OnClicked: func() {
									getUrl := url.Text()
									getConnCount, _ := strconv.Atoi(connCount.Text())
									go mainGet(&getUrl, &getConnCount, &outTE)
								},
							},
						},
					},
					TextEdit{AssignTo: &outTE},
				},
			},
		},
	}.Run()
}

var (
	nPtr       = 4
	urlPtr     = ""
	outFilePtr = "defaultfilename"
)

type download struct {
	url         string
	dlStatus    chan string
	fileName    string
	fileStatus  string
	connections int
	size        int
	client      *http.Client
	outTE       *walk.TextEdit
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

	req.Header.Set("Range", reqRange)

	filename := dl.fileName + "." + strconv.Itoa(i)
	out, _ := os.Create(filename)

	client := dl.client
	resp, _ := client.Do(req)
	writer := io.MultiWriter(out, dl.pb)


	_, err := io.Copy(writer,resp.Body)

	if err != nil {
		dl.outTE.SetText(fmt.Sprintln("Error occured", err))
		dl.dlStatus <- "failed"
		return
	} else {
		//"Success", filename, " written", written))
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

		_, err := io.Copy(out, in)
		in.Close()
		defer os.Remove(infile)
		if err != nil {
			dl.outTE.SetText(fmt.Sprintln("Error occured", err))

		} else {
			dl.outTE.SetText(fmt.Sprintln("file ", infile, "joined"))

		}

	}

	dl.outTE.SetText(fmt.Sprintln("file join complete"))
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
func mainGet(url *string, connCount *int, outTE **walk.TextEdit) {
	cpuCount := runtime.NumCPU()
	i := runtime.GOMAXPROCS(runtime.NumCPU())

	startTime := time.Now()
	flag.Parse()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	dl := download{
		url:         *url,
		fileName:    "defaultfilename",
		connections: *connCount,
		dlStatus:    make(chan string, 1),
		client:      &http.Client{Transport: tr},
		outTE:       *outTE,
	}
	dl.outTE.SetText(fmt.Sprintln("Was using ", i, "CPUs changing to ", cpuCount))
	//nowTime := time.Now().Format("20060102-150405")
	if dl.fileName == "defaultfilename" {
		//dl.fileName = nowTime
		i := strings.LastIndex(dl.url, "/")
		dl.fileName = dl.url[i+1:]
		dl.outTE.SetText(fmt.Sprintln("Filename :", dl.fileName))
	}
	if dl.url == "" {
		reader := bufio.NewReader(os.Stdin)
		outTE.SetText(fmt.Sprintln("Please enter a valid url: "))
		dl.url, _ = reader.ReadString('\n')
	}

	resp, err := dl.client.Get(dl.url)

	if err != nil {

	}

	dl.size = int(resp.ContentLength)
	isPartial := resp.Header.Get("Accept-Ranges")
	if isPartial == "" {
		fmt.Println("Partial downloads not supported")
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
	dl.join()

	speed := (float64(dl.size / 1024)) / timeTaken
	outTE.SetText(fmt.Sprintf("done time taken:  %s , %.4f Speed:kB/s\n", timeTakenS, speed))

}


