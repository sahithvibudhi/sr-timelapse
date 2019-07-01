package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
)

var workspace string
var wg sync.WaitGroup
var fps int32 = int32(*flag.Int("fps", 24, "Frames per Second."))
var storagePath []string

func main() {

	workspace = "/Users/sahithvibudhi/go-workspace/src/github.com/sahithvibudhi/ss-timelapse"
	flag.Parse()
	n := screenshot.NumActiveDisplays()
	wg.Add(n)
	storagePath = make([]string, n)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	fmt.Printf("found %d screen(s).\n", n)

	for i := 0; i < n; i++ {

		go captureScreen(i)

	}

	stopRecording(c)
	wg.Wait()

}

func captureScreen(screenNumber int) {

	defer wg.Done()
	bounds := screenshot.GetDisplayBounds(screenNumber)
	storagePath[screenNumber] = workspace + "/screen-" + strconv.Itoa(screenNumber) + "/"

	for i := 0; ; i++ {

		img, err := screenshot.CaptureRect(bounds)
		checkErr(err)
		os.Mkdir(storagePath[screenNumber], 0755)
		fileName := storagePath[screenNumber] + strconv.Itoa(i) + ".jpeg"
		file, _ := os.Create(fileName)
		defer file.Close()
		jpeg.Encode(file, img, &jpeg.Options{Quality: 50})

		fmt.Printf("#%d : %v \"%s\"\n", screenNumber, bounds, fileName)
		time.Sleep(900 * time.Millisecond)

	}

}

func makeVideo(storagePath string) {

	fmt.Println(storagePath)
	aw, err := mjpeg.New("test.avi", 200, 100, fps)
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

}

func stopRecording(c chan os.Signal) {

	<-c

	for _, path := range storagePath {

		makeVideo(path)

	}

	os.Exit(1)

}

func checkErr(err error) {

	if err != nil {

		panic(err)

	}

}
