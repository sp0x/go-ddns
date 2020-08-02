package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sp0x/go-ddns/dnsUpdate"
	"log"
	"net/http"
)

var appConfig = &Config{}
var updater dnsUpdate.Updater

func main() {
	appConfig.LoadConfig("/etc/dyndns.json")
	nsupdater := dnsUpdate.NewNsUpdater(appConfig.NsupdateBinary)
	nsupdater.DefaultTTL = appConfig.RecordTTL
	nsupdater.Server = appConfig.Server
	nsupdater.Domain = appConfig.Domain
	nsupdater.Zone = appConfig.Zone
	updater = nsupdater
	registerRequestParsers()
	router := mux.NewRouter().StrictSlash(true)
	setupRoutes(router)
	log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func DynUpdate(w http.ResponseWriter, r *http.Request) {
	response := BuildWebserviceResponseFromRequest(r, appConfig)

	if !response.Success {
		if response.Message == "Domain not set" {
			_, _ = w.Write([]byte("notfqdn\n"))
		} else {
			_, _ = w.Write([]byte("badauth\n"))
		}
		return
	}

	for _, domain := range response.Domains {
		result := updater.UpdateRecord(domain, response.Address, response.AddrType)
		if result != "" {
			response.Success = false
			response.Message = result

			_, _ = w.Write([]byte("dnserr\n"))
			return
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)
	_, _ = w.Write([]byte(fmt.Sprintf("good %s\n", response.Address)))
}

func Update(w http.ResponseWriter, r *http.Request) {
	response := BuildWebserviceResponseFromRequest(r, appConfig)

	if !response.Success {
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	for _, domain := range response.Domains {
		result := updater.UpdateRecord(domain, response.Address, response.AddrType)

		if result != "" {
			response.Success = false
			response.Message = result

			_ = json.NewEncoder(w).Encode(response)
			return
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

	_ = json.NewEncoder(w).Encode(response)
}
