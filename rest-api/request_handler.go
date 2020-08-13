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
	dnsReq := WebserviceResponse{}
	dnsRequest := dnsRequestExtractors.Extract(r)
	if dnsRequest == nil {
		return WebserviceResponse{
			Success: false,
			Message: "Invalid dnsReq",
		}
	}
	dnsReq.Domains = strings.Split(dnsRequest.Domain, ",")
	dnsReq.DnsRecordValue = dnsRequest.Address

	if dnsRequest.Secret != appConfig.Secret {
		log.Warn(fmt.Sprintf("Invalid dnsReq credential: %s", dnsRequest.Secret))
		dnsReq.Success = false
		dnsReq.Message = "Invalid Credentials"
		return dnsReq
	}

	for i, domain := range dnsReq.Domains {
		if domain == "" {
			dnsReq.Success = false
			dnsReq.Message = "Host not set"
			log.Warn("Host not set")
			return dnsReq
		}
		dnsReq.Domains[i] = parseFullDomain(domain, appConfig)
	}

	var err error
	dnsReq.DnsRecordValue, dnsReq.AddrType, err = parseRecordValue(dnsReq.DnsRecordValue, r)
	if err != nil {
		dnsReq.Success = false
		dnsReq.Message = err.Error()
		return dnsReq
	}
	// kept in the dnsReq for compatibility reasons
	dnsReq.Host = strings.Join(dnsReq.Domains, ",")
	dnsReq.Success = true
	return dnsReq
}

func parseFullDomain(requiredHostNamePart string, c *config.Config) string {
	if strings.HasSuffix(requiredHostNamePart, "."+c.Domain) {
		return requiredHostNamePart
	}
	return fmt.Sprintf("%s.%s", requiredHostNamePart, c.Domain)
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
