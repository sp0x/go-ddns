package api

import (
	"net/http"
)

type dnsRequestExtractor struct {
	Address func(request *http.Request) string
	Secret  func(request *http.Request) string
	Domain  func(request *http.Request) string
}

type RequestExtractionSet []dnsRequestExtractor

var dnsRequestExtractors RequestExtractionSet

func init() {
	registerRequestParsers()
}

type DnsRequest struct {
	Address string
	Secret  string
	Domain  string
}

func registerRequestParsers() {
	e := dnsRequestExtractors
	e = append(e, dnsRequestExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("addr") },
		Secret:  func(r *http.Request) string { return r.URL.Query().Get("secret") },
		Domain:  func(r *http.Request) string { return r.URL.Query().Get("domain") },
	})
	e = append(e, dnsRequestExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("myip") },
		Secret: func(r *http.Request) string {
			_, sharedSecret, ok := r.BasicAuth()
			if !ok || sharedSecret == "" {
				sharedSecret = r.URL.Query().Get("password")
			}
			return sharedSecret
		},
		Domain: func(r *http.Request) string { return r.URL.Query().Get("hostname") },
	})
	dnsRequestExtractors = e
}

func (e RequestExtractionSet) Extract(r *http.Request) *DnsRequest {
	for _, extractor := range e {
		addr := extractor.Address(r)
		secret := extractor.Secret(r)
		domain := extractor.Domain(r)
		if !(secret == "") {
			return &DnsRequest{
				Address: addr,
				Secret:  secret,
				Domain:  domain,
			}
		}
	}
	return nil
}
