package main

import (
	"fmt"
	"log"
	"net/http"
)

func livenesscheckRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "LIVE")
}

func healthcheckRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HEALTHY")
}

func setupLivenessHealthChecks(channelSummaries *AtomicChannelSummaries) {
	http.HandleFunc("/livenesscheck", livenesscheckRoute)
	http.HandleFunc("/healthcheck", healthcheckRoute)
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, channelSummaries)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
