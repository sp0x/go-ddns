package go_ddns

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/go-ddns/config"
	"github.com/sp0x/go-ddns/dnsUpdate"
	"github.com/sp0x/go-ddns/rest-api/api"
	"net/http"
)

var initializedCloudFunc bool
var cloudFuncConfig = &config.Config{}
var cloudUpdater dnsUpdate.Updater

func initialize() {
	if initializedCloudFunc {
		return
	}
	cloudFuncConfig.Load("")
	cloudUpdater = dnsUpdate.NewUpdater(cloudFuncConfig)
	initializedCloudFunc = true
}

func GoDDns(w http.ResponseWriter, r *http.Request) {
	initialize()
	dnsRequest := api.BuildWebserviceResponseFromRequest(r, cloudFuncConfig)

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
		result, err := cloudUpdater.UpdateRecord(domain, dnsRequest.DnsRecordValue, dnsRequest.AddrType)
		if err != nil {
			dnsRequest.Success = false
			dnsRequest.Message = result
			//_ = json.NewEncoder(w).Encode(dnsRequest)
			log.Errorf("couldn't update dns record for %v[%v]='%v': %v", domain, dnsRequest.AddrType, dnsRequest.DnsRecordValue, err)
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
