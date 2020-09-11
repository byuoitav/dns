/*
package piguys

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/caddyserver/caddy"
)

func init() { plugin.Register("piguys", setupPiGuys) }

func setupPiGuys(c *caddy.Controller) error {
	c.Next()
	if c.NextArg() {
		return plugin.Error("piguys", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return PiGuys{}
	})

	return nil
}
*/
