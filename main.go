package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Orion90/portaudio"
)

var writetime = time.Time{}

func main() {
	http.HandleFunc("/fft", fftHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	go http.ListenAndServe(":8080", nil)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 4096)
	stream, err := portaudio.OpenDefaultStream(2, 0, 44100, len(in), in)
	fft_chan := make(chan []int32, 4096)
	go fftanalyzer(fft_chan)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	if err := stream.Start(); err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case <-time.After(5 * time.Millisecond):
				fft_chan <- in
				break
			}
		}
	}()
	for {
		if err := stream.Read(); err != nil {
			fmt.Println(err)
		}
		select {
		case <-sig:
			return
		default:
		}
	}
	chk(stream.Stop())
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
