package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func Start() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2117", nil)
}

//func main() {
//	Start()
//}
