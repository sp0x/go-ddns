package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/go-ddns/config"
	"net"
	"net/http"
	"strings"

	"github.com/sp0x/go-ddns/rest-api/ipparser"
)

type dnsRequestExtractor struct {
	Address func(request *http.Request) string
	Secret  func(request *http.Request) string
	Domain  func(request *http.Request) string
}

type WebserviceResponse struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	Host           string   `json:"domain"`
	Domains        []string `json:"domains"`
	DnsRecordValue string   `json:"address"`
	AddrType       string   `json:"addr_type"`
}

func parseRecordValue(extractedValue string, r *http.Request) (string, string, error) {
	hostValue := extractedValue
	if hostValue == "" {
		hostValue = getRequestRemoteAddress(r)
	}
	if ipparser.IsIPv4(hostValue) {
		return hostValue, "A", nil
	} else if ipparser.IsIPv6(hostValue) {
		return hostValue, "AAAA", nil
	} else {
		log.Warn(fmt.Sprintf("Invalid address: %s", hostValue))
		return "", "", fmt.Errorf("%s is neither a valid IPv4 nor IPv6 address", hostValue)
	}
}

func BuildWebserviceResponseFromRequest(r *http.Request, appConfig *config.Config) WebserviceResponse {
	response := WebserviceResponse{}
	dnsRequest := dnsRequestExtractors.Extract(r)
	if dnsRequest == nil {
		return WebserviceResponse{
			Success: false,
			Message: "Invalid request",
		}
	}
	response.Domains = strings.Split(dnsRequest.Domain, ",")
	response.DnsRecordValue = dnsRequest.Address

	if dnsRequest.Secret != appConfig.Secret {
		log.Warn(fmt.Sprintf("Invalid request credential: %s", dnsRequest.Secret))
		response.Success = false
		response.Message = "Invalid Credentials"
		return response
	}

	for _, domain := range response.Domains {
		if domain == "" {
			response.Success = false
			response.Message = "Host not set"
			log.Warn("Host not set")
			return response
		}
	}

	var err error
	response.DnsRecordValue, response.AddrType, err = parseRecordValue(response.DnsRecordValue, r)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
		return response
	}
	// kept in the response for compatibility reasons
	response.Host = strings.Join(response.Domains, ",")
	response.Success = true
	return response
}

func getRequestRemoteAddress(r *http.Request) string {
	for _, h := range []string{"X-Real-Ip", "X-Forwarded-For"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || addressIsPrivate(realIP) {
				// bad address, go to next
				continue
			}
			return ip
		}
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

type ipRange struct {
	start net.IP
	end   net.IP
}

func networkIsInRange(r ipRange, ipAddress net.IP) bool {
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

var privateAddressRanges = []ipRange{
	{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

func addressIsPrivate(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateAddressRanges {
			// check if this ip is in a private range
			if networkIsInRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}
