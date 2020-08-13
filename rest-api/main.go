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
	router := mux.NewRouter().StrictSlash(true)
	setupRoutes(router)
	log.Println("Dyndns REST services listening on 0.0.0.0:8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Update(w http.ResponseWriter, r *http.Request) {
	dnsRequest := BuildWebserviceResponseFromRequest(r, appConfig)

	if !dnsRequest.Success {
		if dnsRequest.Message == "Host not set" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("notfqdn\n"))
		} else if dnsRequest.Message == "Invalid request" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("badreq\n"))
		} else {
			w.WriteHeader(403)
			_, _ = w.Write([]byte("badauth\n"))
		}
		return
		//non-router dnsRequest.
		//_ = json.NewEncoder(w).Encode(dnsRequest)
	}

	for _, domain := range dnsRequest.Domains {
		result, err := updater.UpdateRecord(domain, dnsRequest.DnsRecordValue, dnsRequest.AddrType)
		if err != nil {
			dnsRequest.Success = false
			dnsRequest.Message = result
			//_ = json.NewEncoder(w).Encode(dnsRequest)
			log.Errorf("couldn't update dns record: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("dnserr\n"))
			return
		}
		log.Infof("Updated %s record %s=%s with result: %s", dnsRequest.AddrType, domain, dnsRequest.DnsRecordValue, result)
	}
	dnsRequest.Success = true
	dnsRequest.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", dnsRequest.AddrType, dnsRequest.Host, dnsRequest.DnsRecordValue)
	_ = json.NewEncoder(w).Encode(dnsRequest)
}
