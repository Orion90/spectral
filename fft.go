package main

import (
	"log"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan [512]float64)

func fftanalyzer(values chan []int32) {
	for {
		in := <-values
		var buf []float64
		for _, a := range in {
			buf = append(buf, float64(a))
		}
		buffer := fft.FFTReal(buf)
		var fft_buf [512]float64
		for i := 0; i <= 510; i += 2 {
			val := buffer[i] + buffer[i+1]
			fft_buf[i/2] = cmplx.Abs(val)
			if fft_buf[i/2] < 0 {
				log.Fatalln("How the fuck can I can a negative abs-value", i/2, buffer)
			}
		}
		fft_values <- fft_buf
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

type BarData struct {
	Values []ValuePair `json:"values"`
}

type ValuePair struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func fftHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()
	for {
		values := <-fft_values
		var send []int32
		for _, v := range values {
			send = append(send, int32(v)%100)
		}
		if err := c.WriteJSON(send); err != nil {
			log.Fatalln(err)
		}
	}
}
