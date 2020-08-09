package dnsUpdate

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sp0x/go-ddns/dnsUpdate/mocks"
	"google.golang.org/api/dns/v1"
	"testing"
)

func TestGoogleCloudDns_UpdateRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := NewGomegaWithT(t)
	defer ctrl.Finish()
	//config := mocks.NewMock
	dnsUpdater := makeGoogleDns()
	adapter := mocks.NewMockDnsServiceAdapter(ctrl)
	dnsUpdater.serviceAdapter = adapter
	adapter.EXPECT().
		List("zonecom", "one.com", "A").
		Times(1).
		Return(nil, nil)
	exampleChange := &dns.Change{
		Additions: []*dns.ResourceRecordSet{
			{
				Name:    "one.com",
				Rrdatas: []string{"1.1.1.1"},
				Ttl:     300,
				Type:    "A",
			},
		},
	}
	adapter.EXPECT().
		Change("zonecom", gomock.Eq(exampleChange)).
		Times(1).
		Return("ok", nil)
	result, err := dnsUpdater.UpdateRecord("one.com", "1.1.1.1", "A")
	g.Expect(err).To(BeNil())
	g.Expect(result).To(Equal("ok"))
}

func makeGoogleDns() *GoogleCloudDns {
	output := &GoogleCloudDns{project: "project_x", ttl: 300}
	output.ctx = context.Background()
	output.SetZone("zonecom")
	return output
}
