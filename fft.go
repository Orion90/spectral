package main

import (
	"log"
	"math"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan [512]int)

func fftanalyzer(values chan []int32) {
	for {
		in := <-values
		var buf []float64
		for k := 0; k <= len(in)-2; k += 2 {
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
		var send [512]int
		for h, v := range buf {
			send[h] = int(v) % 100
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

func avgInt32(vals []int) int {
	sum := 0
	for _, v := range vals {
		sum += v
	}
	return sum / len(vals)
}

func fftHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()
	for {
		var a []byte
		c.ReadJSON(&a)
		values := <-fft_values
		bs := []BarData{
			BarData{},
		}
		for i, _ := range values[0:64] {
			bs[0].Values = append(bs[0].Values, ValuePair{i, avgInt32(values[i*4 : (i*4 + 4)])})
		}
		c.WriteJSON(bs)
	}
}
