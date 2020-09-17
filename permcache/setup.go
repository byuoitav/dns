package permcache

import (
	"fmt"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	bolt "go.etcd.io/bbolt"
)

func init() {
	plugin.Register(_name, setup)
}

func setup(c *caddy.Controller) error {
	c.Next() // 'permcache'
	if !c.NextArg() {
		return plugin.Error(_name, c.ArgErr())
	}

	path := c.Val()

	if c.NextArg() {
		return plugin.Error(_name, c.ArgErr())
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return plugin.Error(_name, fmt.Errorf("can't open database: %w", err))
	}

	// make sure the bucket exists
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(_bucket)
		return err
	})
	if err != nil {
		return plugin.Error(_name, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &Cache{
			Next: next,
			db:   db,
		}
	})

	return nil
}
