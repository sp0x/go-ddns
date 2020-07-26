package main

import "github.com/gorilla/mux"

func setupRoutes(router *mux.Router) {
	router.HandleFunc("/update", Update).Methods("GET")

	/* DynDNS compatible handlers. Most routers will invoke /nic/update */
	router.HandleFunc("/nic/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v2/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v3/update", DynUpdate).Methods("GET")
}
