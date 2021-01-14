package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type HTTPServer struct {
	Addr string
}

func ServeHTTP(ctx context.Context, s *HTTPServer) (err error) {
	logrus.Infof("Starting prometheus metrics.")
	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.HandleFunc("/events", FindEvents).Methods("GET")
	r.HandleFunc("/events/sample", FindEventsSample).Methods("GET")
	r.HandleFunc("/events/type", FindEventsByType).Methods("GET")

	go func() {
		if err = http.ListenAndServe(s.Addr, r); err != nil {
			logrus.Warnf("http server error: " + err.Error())
		}
	}()
	<-ctx.Done()
	return
}
