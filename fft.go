package main

import (
	"log"
	"math"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan []int64, 64)

func fftanalyzer(values chan []int32) {
	for {
		in := <-values
		var buf [2048]float64
		for k := 0; k <= (len(in)/2)-2; k++ {
			v1 := float64(in[k*2]) * 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(k*2)/(float64(len(in))-1.0)))
			v2 := float64(in[k*2+1]) * 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(k*2)/(float64(len(in))-1.0)))
			buf[k] = v1 + v2
		}
		buffer := fft.FFTReal(buf[0:2048])
		for i, val := range buffer {
			buf[i] = cmplx.Abs(val)
		}
		var sendBuf [64]float64
		sendBuf[1] = avgFloat64(buf[freqToIndex(31):freqToIndex(60)])
		sendBuf[2] = avgFloat64(buf[freqToIndex(61):freqToIndex(100)])
		sendBuf[3] = avgFloat64(buf[freqToIndex(101):freqToIndex(150)])
		sendBuf[4] = avgFloat64(buf[freqToIndex(151):freqToIndex(200)])
		sendBuf[5] = avgFloat64(buf[freqToIndex(201):freqToIndex(250)])
		sendBuf[6] = avgFloat64(buf[freqToIndex(251):freqToIndex(300)])
		sendBuf[7] = avgFloat64(buf[freqToIndex(301):freqToIndex(350)])
		sendBuf[8] = avgFloat64(buf[freqToIndex(351):freqToIndex(400)])
		for n := 9; n < 64; n++ {
			sendBuf[n] = avgFloat64(buf[freqToIndex(351+(250*(n-9))):freqToIndex(500+(250*(n-9)))])
		}
		send := make([]int64, 64)
		x := 8.0
		y := 3.0
		var avg int64
		for h, v := range sendBuf {
			q := v * (math.Log(x) / y)
			x = x + (x)
			send[h] = int64(q)
			avg += send[h]
		}
		avg = avg / int64(len(send))
		for g, va := range send {
			if va > avg {
				send[g] = send[g] / 4
			}
		}
		fft_values <- send
	}
}
func freqToIndex(freq int) int {
	return freq / (44100 / 2048)
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

type BarData struct {
	Values []ValuePair `json:"values"`
}

type ValuePair struct {
	X int   `json:"x"`
	Y int64 `json:"y"`
}

func avgFloat64(vals []float64) float64 {
	sum := float64(0)
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
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
		c.ReadMessage()
		values := <-fft_values
		bs := []BarData{
			BarData{},
		}
		for i, _ := range values[0:64] {
			bs[0].Values = append(bs[0].Values, ValuePair{i, values[i]})
		}
		if err := c.WriteJSON(bs); err != nil {
			return
		}
	}
}
