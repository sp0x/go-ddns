package dnsUpdate

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
	"log"
	"os"
)

type GoogleCloudDns struct {
	ctx            context.Context
	service        *dns.Service
	serviceAdapter DnsServiceAdapter
	zone           string
	project        string
	ttl            int64
}

//go:generate mockgen -destination=mocks/mock_dnsServiceAdapter.go -package=mocks . DnsServiceAdapter
type DnsServiceAdapter interface {
	List(zone string, domain string, recType string) ([]*dns.ResourceRecordSet, error)
	Change(zone string, change *dns.Change) (string, error)
}

type GoogleServiceAdapter struct {
	service *dns.Service
	project string
	ctx     context.Context
}

func (gs *GoogleServiceAdapter) List(zone string, domain string, recType string) ([]*dns.ResourceRecordSet, error) {
	records, err := gs.service.ResourceRecordSets.List(gs.project, zone).
		Name(domain).
		Type(recType).
		Do()
	if err != nil {
		return nil, err
	}
	return records.Rrsets, nil
}

func (gs *GoogleServiceAdapter) Change(zone string, change *dns.Change) (string, error) {
	resp, err := gs.service.Changes.Create(gs.project, zone, change).Context(gs.ctx).Do()
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

func NewGoogleDns(projectName string) *GoogleCloudDns {
	output := &GoogleCloudDns{project: projectName, ttl: 300}
	output.ctx = context.Background()
	dnsService, err := dns.NewService(output.ctx, option.WithScopes(dns.CloudPlatformScope))
	if err != nil {
		fmt.Printf("couldn't initialize google dns service: %v", err)
		os.Exit(1)
	}
	output.service = dnsService
	output.serviceAdapter = &GoogleServiceAdapter{dnsService, projectName, output.ctx}
	return output
}

func (ns *GoogleCloudDns) SetZone(zoneName string) {
	ns.zone = zoneName
}

func (ns *GoogleCloudDns) ListZone() ([]*dns.ResourceRecordSet, error) {
	if ns.zone == "" {
		return nil, errors.New("zone is empty")
	}
	var output []*dns.ResourceRecordSet
	req := ns.service.Changes.List(ns.project, ns.zone)
	if err := req.Pages(ns.ctx, func(page *dns.ChangesListResponse) error {
		for _, change := range page.Changes {
			output = append(output, change.Additions...)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return output, nil
}

func (ns *GoogleCloudDns) ListRecordsWithName(name string, recordType string) ([]*dns.ResourceRecordSet, error) {
	records, err := ns.serviceAdapter.List(ns.zone, name, recordType)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (ns *GoogleCloudDns) HasAFor(name string) (bool, error) {
	records, err := ns.ListRecordsWithName(name, "A")
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

func (ns *GoogleCloudDns) UpdateRecord(domain string, value string, recordType string) (string, error) {
	// see https://cloud.google.com/dns/docs/records/json-record for information
	var record = &dns.ResourceRecordSet{
		Type:    recordType,
		Name:    domain,
		Ttl:     ns.ttl,
		Rrdatas: []string{value},
	}
	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{record},
	}
	//We have to delete any existing records
	existingRecords, err := ns.ListRecordsWithName(domain, recordType)
	if err != nil {
		return "", err
	}
	change.Deletions = existingRecords

	resp, err := ns.serviceAdapter.Change(ns.zone, change)
	if err != nil {
		return "", err
	}
	return resp, nil
}
