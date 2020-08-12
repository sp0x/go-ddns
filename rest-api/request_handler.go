package main

import (
	"bytes"
	"errors"
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
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	Domain   string   `json:"domain"`
	Domains  []string `json:"domains"`
	Address  string   `json:"address"`
	AddrType string   `json:"addr_type"`
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
	response.Address = dnsRequest.Address

	if dnsRequest.Secret != appConfig.Secret {
		log.Warn(fmt.Sprintf("Invalid request credential: %s", dnsRequest.Secret))
		response.Success = false
		response.Message = "Invalid Credentials"
		return response
	}

	for _, domain := range response.Domains {
		if domain == "" {
			response.Success = false
			response.Message = "Domain not set"
			log.Warn("Domain not set")
			return response
		}
	}

	// kept in the response for compatibility reasons
	response.Domain = strings.Join(response.Domains, ",")

	if ipparser.IsIPv4(response.Address) {
		response.AddrType = "A"
	} else if ipparser.IsIPv6(response.Address) {
		response.AddrType = "AAAA"
	} else {
		var ip string
		var err error

		ip, err = getRequestRemoteAddress(r)
		if ip == "" {
			ip, _, err = net.SplitHostPort(r.RemoteAddr)
		}

		if err != nil {
			response.Success = false
			response.Message = fmt.Sprintf("%q is neither a valid IPv4 nor IPv6 address", r.RemoteAddr)
			log.Warn(fmt.Sprintf("Invalid address: %q", r.RemoteAddr))
			return response
		}
		if ipparser.IsIPv4(ip) {
			response.AddrType = "A"
		} else if ipparser.IsIPv6(ip) {
			response.AddrType = "AAAA"
		} else {
			response.Success = false
			response.Message = fmt.Sprintf("%s is neither a valid IPv4 nor IPv6 address", response.Address)
			log.Warn(fmt.Sprintf("Invalid address: %s", response.Address))
			return response
		}

		response.Address = ip
	}
	response.Success = true
	return response
}

func getRequestRemoteAddress(r *http.Request) (string, error) {
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
			return ip, nil
		}
	}
	return "", errors.New("no match")
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
