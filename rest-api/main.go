package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sp0x/docker-ddns/dnsUpdate"
	"log"
	"net/http"
)

var appConfig = &Config{}
var updater dnsUpdate.Updater

func main() {
	appConfig.LoadConfig("/etc/dyndns.json")

	router := mux.NewRouter().StrictSlash(true)
	setupRoutes(router)
	nsupdater := dnsUpdate.NewNsUpdater(appConfig.NsupdateBinary)
	nsupdater.DefaultTTL = appConfig.RecordTTL
	nsupdater.Server = appConfig.Server
	nsupdater.Domain = appConfig.Domain
	nsupdater.Zone = appConfig.Zone
	updater = nsupdater
	log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func DynUpdate(w http.ResponseWriter, r *http.Request) {
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("myip") },
		Secret: func(r *http.Request) string {
			_, sharedSecret, ok := r.BasicAuth()
			if !ok || sharedSecret == "" {
				sharedSecret = r.URL.Query().Get("password")
			}

			return sharedSecret
		},
		Domain: func(r *http.Request) string { return r.URL.Query().Get("hostname") },
	}
	response := BuildWebserviceResponseFromRequest(r, appConfig, extractor)

	if response.Success == false {
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
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("addr") },
		Secret:  func(r *http.Request) string { return r.URL.Query().Get("secret") },
		Domain:  func(r *http.Request) string { return r.URL.Query().Get("domain") },
	}
	response := BuildWebserviceResponseFromRequest(r, appConfig, extractor)

	if response.Success == false {
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
