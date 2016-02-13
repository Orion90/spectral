package main

import (
	"log"
	"math"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan [32]float64)

func fftanalyzer(values chan []int32) {
	for {
		in := <-values
		var buf []float64
		for _, a := range in {
			buf = append(buf, float64(a))
		}
		buffer := fft.FFTReal(buf)
		var fft_buf [32]float64
		for i := 0; i <= 62; i += 2 {
			val := buffer[i] + buffer[i+1]
			fft_buf[i/2] = (cmplx.Abs(val) * math.Pow(10, -8))
		}
		fft_values <- fft_buf
	}
}

var upgrader = websocket.Upgrader{} // use default options

func fftHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()
	for {
		values := <-fft_values
		if err := c.WriteJSON(values); err != nil {
			log.Fatalln(err)
		}
	}
}
