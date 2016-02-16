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
		sendBuf[0] = fftAvg(buf[0:len(buf)-1], 0, 30)
		sendBuf[1] = fftAvg(buf[0:len(buf)-1], 31, 60)
		sendBuf[2] = fftAvg(buf[0:len(buf)-1], 61, 100)
		sendBuf[3] = fftAvg(buf[0:len(buf)-1], 101, 150)
		sendBuf[4] = fftAvg(buf[0:len(buf)-1], 151, 200)
		sendBuf[5] = fftAvg(buf[0:len(buf)-1], 201, 250)
		sendBuf[6] = fftAvg(buf[0:len(buf)-1], 251, 300)
		sendBuf[7] = fftAvg(buf[0:len(buf)-1], 301, 350)
		sendBuf[8] = fftAvg(buf[0:len(buf)-1], 351, 400)
		for n := 9; n < 64; n++ {
			sendBuf[n] = fftAvg(buf[0:len(buf)-1], (351 + (250 * (n - 9))), (500 + (250 * (n - 9))))
		}
		send := make([]int64, 64)
		x := 8.0
		y := 3.0
		for h, v := range sendBuf {
			q := v * (math.Log(x) / y)
			x = x + (x)
			send[h] = int64(q) / math.MaxInt32
		}
		fft_values <- send
	}
}
func fftAvg(values []float64, start, stop int) float64 {
	sum := 0.0
	for i := start; i <= stop; i++ {
		sum += values[freqToIndex(i)]
	}
	return sum / float64(start-stop+1)
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
