package secret_manager

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

var isAlive bool
var serverHttpPort int

func startHttpServer() {
	http.HandleFunc("/liveness", liveness)

	glog.Info("Http server listening on " + fmt.Sprintf("%d", serverHttpPort))

	go func() {
		err := http.ListenAndServe(":"+fmt.Sprintf("%d", serverHttpPort), nil)

		if err != nil {
			glog.Exit("Failed to start HTTP server")
		}
	}()
}

func liveness(w http.ResponseWriter, req *http.Request) {
	glog.V(2).Info("Received liveness request")
	if !isAlive {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Service is not ready yet"))

		glog.V(2).Info("Returning a 500 status code for the liveness request")

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	glog.V(2).Info("Returning a 200 status code for the liveness request")
}
