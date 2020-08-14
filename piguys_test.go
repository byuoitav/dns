package piguys

import (
	"fmt"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestPiGuys(t *testing.T) {
	p := PiGuys{}
	if p.Name() != name {
		t.Errorf("expected plugin name: %s, got %s", p.Name(), name)
	}
	tests := []struct {

	}

	ctx := context.TODO()
	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("ITB-1010-CP1.byu.edu"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{RemoteIP: ""})
	code, err := p.ServeDNS(ctx, rec, req)
	if err != nil {
		fmt.Errorf("oh no there was an error")
	}
}
