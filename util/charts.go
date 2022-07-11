package util

import (
	"bytes"
	"github.com/wcharczuk/go-chart"
	"log"
	"os"
	"time"
)

// ChartsData jbzhou5 绘制图表所用的值
type ChartsData struct {
	Val   float64   // 纵坐标值
	XTime time.Time // 横坐标时间
}

func main() {
	graph := chart.Chart{
		Title: "性能评估",
		XAxis: chart.XAxis{
			Name: "时间",
			ValueFormatter: func(v interface{}) string {
				return v.(time.Time).Format("2006-01-02 15:05:05.0000")
			},
		},
		YAxis: chart.YAxis{
			Name: "时间/ms",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: []float64{1.0, 2.0, 3.0, 4.0},
				YValues: []float64{1.0, 2.0, 3.0, 4.0},
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	_ = graph.Render(chart.PNG, buffer)
	out, err := os.Create("out.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	out.Write(buffer.Bytes())
}
