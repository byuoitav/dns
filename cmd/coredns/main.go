package main

import (
	_ "github.com/byuoitav/dns/permcache"
	_ "github.com/coredns/coredns/core/plugin"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

func init() {
	added := make(map[string]bool)

	// insert permcache after cache
	for i, name := range dnsserver.Directives {
		if name == "cache" {
			dnsserver.Directives = append(dnsserver.Directives[:i], append([]string{"permcache"}, dnsserver.Directives[i:]...)...)
			added["cache"] = true
		}
	}

	if !added["cache"] {
		panic("plugins not added to dnsserver.Directives")
	}
}

func main() {
	coremain.Run()
}
