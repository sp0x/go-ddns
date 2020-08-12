package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/go-ddns/config"
	"github.com/sp0x/go-ddns/dnsUpdate"
	"net/http"
)

var appConfig = &config.Config{}
var updater dnsUpdate.Updater

func main() {
	appConfig.Load("/etc/goddns.yml")
	updater = dnsUpdate.NewUpdater(appConfig)
	//updater.DefaultTTL = appConfig.RecordTTL
	//updater.Server = appConfig.Server
	//updater.Domain = appConfig.Domain
	//updater.Zone = appConfig.Zone
	router := mux.NewRouter().StrictSlash(true)
	setupRoutes(router)
	log.Println("Dyndns REST services listening on 0.0.0.0:8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Update(w http.ResponseWriter, r *http.Request) {
	response := BuildWebserviceResponseFromRequest(r, appConfig)

	if !response.Success {
		if response.Message == "Domain not set" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("notfqdn\n"))
		} else if response.Message == "Invalid request" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("badreq\n"))
		} else {
			w.WriteHeader(403)
			_, _ = w.Write([]byte("badauth\n"))
		}
		return
		//non-router response.
		//_ = json.NewEncoder(w).Encode(response)
	}

	for _, domain := range response.Domains {
		result, err := updater.UpdateRecord(domain, response.Address, response.AddrType)
		if err != nil {
			response.Success = false
			response.Message = result
			//_ = json.NewEncoder(w).Encode(response)
			log.Errorf("couldn't update dns record: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("dnserr\n"))
			return
		}
		if result != "" {
			response.Success = false
			response.Message = result
			//_ = json.NewEncoder(w).Encode(response)
			_, _ = w.Write([]byte("dnserr\n"))
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)
	_ = json.NewEncoder(w).Encode(response)
}
