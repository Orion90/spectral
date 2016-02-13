package main

import (
	"log"
	"math"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan []float64)

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
		fft_values <- buf
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
		var bar BarData
		for i, v := range values[2:512] {
			if i == 0 {
				v = 0
			}
			if v < 0 {
				log.Println("Negative value..", i, v)
			}
			bar.Values = append(bar.Values, ValuePair{i, int(v) % 100})
		}
		bs := []BarData{bar}
		if err := c.WriteJSON(bs); err != nil {
			log.Fatalln(err)
		}
	}
}
