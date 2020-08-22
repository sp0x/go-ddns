package dnsUpdate

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sp0x/go-ddns/dnsUpdate/mocks"
	"google.golang.org/api/dns/v1"
	"strings"
	"testing"
)

func TestGoogleCloudDns_UpdateRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := NewGomegaWithT(t)
	defer ctrl.Finish()
	dnsUpdater := makeGoogleDns()
	adapter := mocks.NewMockDnsServiceAdapter(ctrl)
	dnsUpdater.serviceAdapter = adapter
	adapter.EXPECT().
		List("zonecom", "one.com.", "A").
		Times(1).
		Return(nil, nil)
	exampleChange := &dns.Change{
		Additions: []*dns.ResourceRecordSet{
			{
				Name:    "one.com.",
				Rrdatas: []string{"1.1.1.1"},
				Ttl:     300,
				Type:    "A",
			},
		},
	}
	adapter.EXPECT().
		Change("zonecom", withChange(exampleChange)).
		Times(1).
		Return("ok", nil)
	result, err := dnsUpdater.UpdateRecord("one.com.", "1.1.1.1", "A")
	g.Expect(err).To(BeNil())
	g.Expect(result).To(Equal("ok"))
}

func makeGoogleDns() *GoogleCloudDns {
	output := &GoogleCloudDns{project: "project_x", ttl: 300}
	output.ctx = context.Background()
	output.SetZone("zonecom")
	return output
}

type changeMatch struct{ change *dns.Change }

func withChange(t *dns.Change) gomock.Matcher {
	return &changeMatch{t}
}
func (o *changeMatch) Matches(x interface{}) bool {
	data, _ := x.(*dns.Change)
	for i, expected := range o.change.Additions {
		got := data.Additions[i]
		if got.Name != expected.Name {
			return false
		}
		if got.Type != expected.Type {
			return false
		}
		if got.Ttl != expected.Ttl {
			return false
		}
		if strings.Join(got.Rrdatas, ",") != strings.Join(expected.Rrdatas, ",") {
			return false
		}
	}
	return true
}

func (o *changeMatch) String() string {
	return spew.Sdump(o.change.Additions)
}
