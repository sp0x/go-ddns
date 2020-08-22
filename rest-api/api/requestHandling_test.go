package api

import (
	"github.com/sp0x/go-ddns/config"
	"net/http"
	"testing"
)

func TestBuildWebserviceResponseFromRequestToReturnValidObject(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"
	appConfig.Domain = "example.com"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Host != "foo.example.com" {
		t.Fatalf("Expected WebserviceResponse.Host to be foo.example.com, but was `%v`", result.Host)
	}

	if result.DnsRecordValue != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.DnsRecordValue to be 1.2.3.4")
	}

	if result.AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestWithXRealIPHeaderToReturnValidObject(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"
	appConfig.Domain = "example.com"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo", nil)
	req.Header.Add("X-Real-Ip", "1.2.3.4")
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Host != "foo.example.com" {
		t.Fatalf("Expected WebserviceResponse.Host to be foo.example.com")
	}

	if result.DnsRecordValue != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.DnsRecordValue to be 1.2.3.4")
	}

	if result.AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestWithXForwardedForHeaderToReturnValidObject(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"
	appConfig.Domain = "example.com"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.4")
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Host != "foo.example.com" {
		t.Fatalf("Expected WebserviceResponse.Host to be foo.example.com")
	}

	if result.DnsRecordValue != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.DnsRecordValue to be 1.2.3.4")
	}

	if result.AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoSecretIsGiven(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/update", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidSecretIsGiven(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/update?secret=foo", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoDomainIsGiven(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/update?secret=changeme", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestWithMultipleDomains(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"
	appConfig.Domain = "example.com"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo,bar&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if len(result.Domains) != 2 {
		t.Fatalf("Expected WebserviceResponse.Domains length to be 2")
	}

	if result.Domains[0] != "foo.example.com" {
		t.Fatalf("Expected WebserviceResponse.Domains[0] to equal 'foo.example.com'")
	}

	if result.Domains[1] != "bar.example.com" {
		t.Fatalf("Expected WebserviceResponse.Domains[1] to equal 'bar.example.com'")
	}
}

func TestBuildWebserviceResponseFromRequestWithMalformedMultipleDomains(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo,&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoAddressIsGiven(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("POST", "/update?secret=changeme&domain=foo", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidAddressIsGiven(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo&addr=1.41:2", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnValidObjectWithDynExtractor(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"
	appConfig.Domain = "example.com"

	req, _ := http.NewRequest("GET", "/nic/update?hostname=foo&myip=1.2.3.4", nil)
	req.Header.Add("Authorization", "Basic dXNlcm5hbWU6Y2hhbmdlbWU=") // This is the base-64 encoded value of "username:changeme"

	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Host != "foo.example.com" {
		t.Fatalf("Expected WebserviceResponse.Host to be foo")
	}

	if result.DnsRecordValue != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.DnsRecordValue to be 1.2.3.4")
	}

	if result.AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoSecretIsGivenWithDynExtractor(t *testing.T) {
	var appConfig = &config.Config{}
	appConfig.Secret = "changeme"

	req, _ := http.NewRequest("GET", "/nic/update", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}
