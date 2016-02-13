package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Orion90/portaudio"
)

var (
	o = flag.String("o", "", "Save to disk as file")
)

func main() {
	http.HandleFunc("/fft", fftHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	go http.ListenAndServe(":8080", nil)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	f, err := os.Create("test.aiff")
	if err != nil {
		log.Fatalln("Couldn't create file'")
	}
	defer f.Close()
	// form chunk
	_, err = f.WriteString("FORM")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(0))) //total bytes
	_, err = f.WriteString("AIFF")
	chk(err)

	// common chunk
	_, err = f.WriteString("COMM")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(18)))                  //size
	chk(binary.Write(f, binary.BigEndian, int16(2)))                   //channels
	chk(binary.Write(f, binary.BigEndian, int32(0)))                   //number of samples
	chk(binary.Write(f, binary.BigEndian, int16(32)))                  //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	chk(err)

	// sound chunk
	_, err = f.WriteString("SSND")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(0))) //size
	chk(binary.Write(f, binary.BigEndian, int32(0))) //offset
	chk(binary.Write(f, binary.BigEndian, int32(0))) //block
	nSamples := 0
	defer func() {
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		_, err = f.Seek(4, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(totalBytes)))
		_, err = f.Seek(22, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(nSamples)))
		_, err = f.Seek(42, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(4*nSamples+8)))
		chk(f.Close())
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(2, 0, 44100, len(in), in)
	fft_chan := make(chan []int32, 64)
	go fftanalyzer(fft_chan)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	if err := stream.Start(); err != nil {
		panic(err)
	}
	for {
		if err := stream.Read(); err != nil {
			fmt.Println(err)
		}
		if err := binary.Write(f, binary.BigEndian, in); err != nil {
			panic(err)
		}
		fft_chan <- in
		nSamples += len(in)
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
