package dnsUpdate

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type Updater interface {
	UpdateRecord(domain string, ipaddr string, addrType string) string
}

type NSUpdate struct {
	Server     string
	Zone       string
	Domain     string
	DefaultTTL int
	binary     string
}

func NewNsUpdater(binary string) *NSUpdate {
	ns := &NSUpdate{}
	ns.DefaultTTL = 300
	ns.binary = binary
	return ns
}

func (ns *NSUpdate) UpdateRecord(domain string, ipaddr string, recordType string) string {
	log.Println(fmt.Sprintf("%s record update request: %s -> %s", recordType, domain, ipaddr))

	f, err := ioutil.TempFile(os.TempDir(), "dyndns")
	if err != nil {
		return err.Error()
	}

	defer func() {
		_ = os.Remove(f.Name())
	}()
	w := bufio.NewWriter(f)
	nsDomain := ns.Domain
	ttl := ns.DefaultTTL
	_, _ = w.WriteString(fmt.Sprintf("server %s\n", ns.Server))
	_, _ = w.WriteString(fmt.Sprintf("zone %s\n", ns.Zone))
	_, _ = w.WriteString(fmt.Sprintf("update delete %s.%s %s\n", domain, nsDomain, recordType))
	_, _ = w.WriteString(fmt.Sprintf("update add %s.%s %v %s %s\n", domain, nsDomain, ttl, recordType, ipaddr))
	_, _ = w.WriteString("send\n")

	_ = w.Flush()
	_ = f.Close()

	cmd := exec.Command(ns.binary, f.Name())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err.Error() + ": " + stderr.String()
	}
	return out.String()
}
