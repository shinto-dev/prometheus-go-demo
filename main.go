package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"time"
)

var RequestCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "go_app_requests_count",
	Help: "Total app requests count.",
})

var RequestInProgress = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "app_requests_in_progress",
	Help: "Requests in progress",
})

// tracks the request response time for each api
var RequestRespondTime = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "app_response_latency_seconds",
	Help: "Response latency in seconds",
}, []string{"path"})

var RequestRespondTimeHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "app_response_latency_histogram_seconds",
	Help: "Response latency in seconds",
}, []string{"path"})

func routeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		next.ServeHTTP(w, r)
		timeTaken := time.Since(startTime)
		RequestRespondTime.WithLabelValues(path).Observe(timeTaken.Seconds())
		RequestRespondTimeHistogram.WithLabelValues(path).Observe(timeTaken.Seconds())
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/birthday/{name}", func(w http.ResponseWriter, r *http.Request) {
		RequestInProgress.Inc()
		defer RequestInProgress.Dec()
		RequestCount.Inc()

		time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
		name := mux.Vars(r)["name"]
		greeting := fmt.Sprintf("Happy Birthday %s", name)
		w.Write([]byte(greeting))
	}).Methods("GET")

	router.Path("/metrics").Handler(promhttp.Handler())
	router.Use(routeMiddleware)

	fmt.Println("starting server")
	err := http.ListenAndServe(":8001", router)
	if err != nil {
		fmt.Println(err)
	}
}
