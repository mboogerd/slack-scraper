package main

import (
	"fmt"
	"log"
	"net/http"
)

func livenessRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "LIVE")
}

func readinessRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HEALTHY")
}

func setupHealthChecks(channelSummaries *AtomicChannelSummaries) {
	http.HandleFunc("/livenesscheck", livenessRoute)
	http.HandleFunc("/readinesscheck", readinessRoute)
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, channelSummaries)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
