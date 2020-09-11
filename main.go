package main

import (
	_ "github.com/byuoitav/dns/permcache"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

var directives = []string{
	"permcache",
	"whoami",
	"log",
}

func init() {
	dnsserver.Directives = directives
}

func main() {
	coremain.Run()
}
