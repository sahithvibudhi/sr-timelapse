package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
)

var workspace string
var wg sync.WaitGroup
var fpsPtr int = *flag.Int("fps", 24, "Frames per Second.")
var fps int32 = int32(fpsPtr)
var intervalPtr int = *flag.Int("interval", 600, "Interval.")
var interval time.Duration = time.Duration(intervalPtr)
var recordingName string = *flag.String("recordingName", "tmp", "Recording Name.")
var storagePath []string

func main() {

	flag.Parse()
	fmt.Println(fps, interval)
	n := screenshot.NumActiveDisplays()
	storagePath = make([]string, n)
	workspace = Workspace(recordingName)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("found %d screen(s).\n", n)
	for i := 0; i < n; i++ {

		go captureScreen(i)

	}

	stopRecording(c)

}

func captureScreen(screenNumber int) {

	bounds := screenshot.GetDisplayBounds(screenNumber)
	storagePath[screenNumber] = workspace + "/screen-" + strconv.Itoa(screenNumber) + "/"
	err := os.MkdirAll(storagePath[screenNumber], 0755)
	checkErr(err)

	for i := 0; ; i++ {

		img, err := screenshot.CaptureRect(bounds)
		checkErr(err)
		fileName := storagePath[screenNumber] + strconv.Itoa(i) + ".jpeg"
		file, _ := os.Create(fileName)
		defer file.Close()
		jpeg.Encode(file, img, &jpeg.Options{Quality: 50})

		fmt.Printf("#%d : %v \"%s\"\n", screenNumber, bounds, fileName)
		time.Sleep(interval * time.Millisecond)

	}

}

func makeVideo(index int, storagePath string) {

	fmt.Println(storagePath)
	filePath := workspace + "/" + recordingName + "screen-" + strconv.Itoa(index) + ".avi"
	aw, err := mjpeg.New(filePath, 200, 100, fps)
	checkErr(err)

	files, err := ioutil.ReadDir(storagePath)
	checkErr(err)
	fileCount := len(files)

	for i, _ := range files {

		data, err := ioutil.ReadFile(storagePath + strconv.Itoa(i) + ".jpeg")
		fmt.Printf("\r Compiled %d/%d images\n", i, fileCount)
		checkErr(err)
		checkErr(aw.AddFrame(data))

	}

	checkErr(aw.Close())

	fmt.Println("Your file is at " + filePath)

}

func stopRecording(c chan os.Signal) {

	<-c

	for i, path := range storagePath {

		makeVideo(i, path)

	}

	os.Exit(1)

}

func Workspace(recordingName string) string {

	var workspace string

	switch _os := runtime.GOOS; _os {
	case "darwin":
		workspace = "/opt/lib/sr-timelapse/" + recordingName
	case "linux":
		workspace = "/opt/lib/sr-timelapse/" + recordingName
	case "windows":
		workspace = "C://sr-timelapse/" + recordingName
	default:
		// freebsd, openbsd,
		// plan9, windows...
		fmt.Printf("sr-timelapse do not support %s yet\n", _os)
		os.Exit(0)
	}

	return workspace

}

func checkErr(err error) {

	if err != nil {

		panic(err)

	}

}
