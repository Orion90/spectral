package main

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mjibson/go-dsp/fft"
)

var fft_values = make(chan [64]int)

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
		/*		sendBuf[0] = 0.7 * avgFloat64(buf[freqToIndex(0):freqToIndex(30)])
				sendBuf[1] = 0.7 * avgFloat64(buf[freqToIndex(31):freqToIndex(60)])
				sendBuf[2] = 0.7 * avgFloat64(buf[freqToIndex(61):freqToIndex(100)])
				sendBuf[3] = 0.7 * avgFloat64(buf[freqToIndex(101):freqToIndex(150)])
				sendBuf[4] = 0.7 * avgFloat64(buf[freqToIndex(151):freqToIndex(200)])
				sendBuf[5] = 0.7 * avgFloat64(buf[freqToIndex(201):freqToIndex(250)])
				sendBuf[6] = 0.7 * avgFloat64(buf[freqToIndex(251):freqToIndex(300)])
				sendBuf[7] = 0.7 * avgFloat64(buf[freqToIndex(301):freqToIndex(350)])
				sendBuf[8] = 0.7 * avgFloat64(buf[freqToIndex(351):freqToIndex(400)])
				for n := 9; n < 64; n++ {
					sendBuf[n] = avgFloat64(buf[freqToIndex(351+(250*(n-9))):freqToIndex(500+(250*(n-9)))])
				}
		*/
		sendBuf[0] = avgFloat64(buf[0:1])
		sendBuf[1] = avgFloat64(buf[1:2])
		sendBuf[2] = avgFloat64(buf[2:4])
		sendBuf[3] = avgFloat64(buf[4:7])
		sendBuf[4] = avgFloat64(buf[7:9])
		sendBuf[5] = avgFloat64(buf[9:11])
		sendBuf[6] = avgFloat64(buf[11:14])
		sendBuf[7] = avgFloat64(buf[14:16])
		sendBuf[8] = avgFloat64(buf[16:19])
		sendBuf[9] = avgFloat64(buf[16:23])
		sendBuf[10] = avgFloat64(buf[28:35])
		sendBuf[11] = avgFloat64(buf[40:47])
		sendBuf[12] = avgFloat64(buf[52:59])
		sendBuf[13] = avgFloat64(buf[64:71])
		sendBuf[14] = avgFloat64(buf[76:83])
		sendBuf[15] = avgFloat64(buf[88:95])
		sendBuf[16] = avgFloat64(buf[100:107])
		sendBuf[17] = avgFloat64(buf[111:119])
		sendBuf[18] = avgFloat64(buf[123:130])
		sendBuf[19] = avgFloat64(buf[135:142])
		sendBuf[20] = avgFloat64(buf[147:154])
		sendBuf[21] = avgFloat64(buf[159:166])
		sendBuf[22] = avgFloat64(buf[171:178])
		sendBuf[23] = avgFloat64(buf[183:190])
		sendBuf[24] = avgFloat64(buf[195:202])
		sendBuf[25] = avgFloat64(buf[207:214])
		sendBuf[26] = avgFloat64(buf[219:226])
		sendBuf[27] = avgFloat64(buf[231:238])
		sendBuf[28] = avgFloat64(buf[242:250])
		sendBuf[29] = avgFloat64(buf[254:261])
		sendBuf[30] = avgFloat64(buf[266:273])
		sendBuf[31] = avgFloat64(buf[278:285])
		sendBuf[32] = avgFloat64(buf[290:297])
		sendBuf[33] = avgFloat64(buf[302:309])
		sendBuf[34] = avgFloat64(buf[314:321])
		sendBuf[35] = avgFloat64(buf[326:333])
		sendBuf[36] = avgFloat64(buf[338:345])
		sendBuf[37] = avgFloat64(buf[350:357])
		sendBuf[38] = avgFloat64(buf[361:369])
		sendBuf[39] = avgFloat64(buf[373:380])
		sendBuf[40] = avgFloat64(buf[385:392])
		sendBuf[41] = avgFloat64(buf[397:404])
		sendBuf[42] = avgFloat64(buf[409:416])
		sendBuf[43] = avgFloat64(buf[421:428])
		sendBuf[44] = avgFloat64(buf[433:440])
		sendBuf[45] = avgFloat64(buf[445:452])
		sendBuf[46] = avgFloat64(buf[457:464])
		sendBuf[47] = avgFloat64(buf[469:476])
		sendBuf[48] = avgFloat64(buf[481:488])
		sendBuf[49] = avgFloat64(buf[492:500])
		sendBuf[50] = avgFloat64(buf[504:511])
		sendBuf[51] = avgFloat64(buf[516:523])
		sendBuf[52] = avgFloat64(buf[528:535])
		sendBuf[53] = avgFloat64(buf[540:547])
		sendBuf[54] = avgFloat64(buf[552:559])
		sendBuf[55] = avgFloat64(buf[564:571])
		sendBuf[56] = avgFloat64(buf[576:583])
		sendBuf[57] = avgFloat64(buf[588:595])
		sendBuf[58] = avgFloat64(buf[600:607])
		sendBuf[59] = avgFloat64(buf[611:619])
		sendBuf[60] = avgFloat64(buf[623:630])
		sendBuf[61] = avgFloat64(buf[635:642])
		sendBuf[62] = avgFloat64(buf[647:654])
		sendBuf[63] = avgFloat64(buf[659:666])
		var send [64]int
		x := 8.0
		y := 3.0
		for h, v := range sendBuf {
			q := v * (math.Log(x) / y)
			x = x + (x)
			send[h] = int(q)
		}
		fft_values <- send
	}
}
func freqToIndex(freq int) int {
	fmt.Println(freq / (44100 / 2048))
	return freq / (44100 / 2048)
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

type BarData struct {
	Values []ValuePair `json:"values"`
}

type ValuePair struct {
	X int `json:"x"`
	Y int `json:"y"`
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
		c.WriteJSON(bs)
	}
}
