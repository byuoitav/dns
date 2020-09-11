package main

import (
	_ "github.com/byuoitav/dns/permcache"
	_ "github.com/coredns/coredns/core/plugin"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

func init() {
	// insert permcache after cache
	for i, name := range dnsserver.Directives {
		if name == "cache" {
			dnsserver.Directives = append(dnsserver.Directives[:i], append([]string{"permcache"}, dnsserver.Directives[i:]...)...)
			return
		}
	}

	panic("cache plugin not found in dnsserver.Directives")
}

func main() {
	coremain.Run()
}
