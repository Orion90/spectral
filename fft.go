package main

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan []int, 1024)

func fftanalyzer(values chan []int32) {
	for {
		in := <-values
		var buf []float64
		for k := 0; k <= len(in)-2; k++ {
			v1 := float64(in[k]) * 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(k)/(float64(len(in))-1.0)))
			v2 := 0.0 //float64(in[k+1]) * 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(k)/(float64(len(in))-1.0)))
			buf = append(buf, v1+v2)
		}
		buffer := fft.FFTReal(buf)
		for i, val := range buffer {

			buf[i] = cmplx.Abs(val)
			if buf[i] < 0 {
				log.Fatalln("How the fuck can I can a negative abs-value", i/2, buffer)
			}
		}
		var send []int
		for _, v := range buf {
			send = append(send, int(v)%100)
		}
		fft_values <- send
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
		now := time.Now()
		var a []byte
		c.ReadJSON(&a)
		values := <-fft_values
		c.WriteJSON(values)
		fmt.Println(time.Since(now))
	}
}
