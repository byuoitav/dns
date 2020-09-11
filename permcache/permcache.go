package permcache

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	bolt "go.etcd.io/bbolt"
)

const (
	_name = "permcache"
)

var (
	_bucket = []byte(_name)
)

type cache struct {
	next plugin.Handler
	db   *bolt.DB
}

func (c cache) Name() string { return _name }

func (c cache) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Printf("PLUGIN PERM CACHE\n")
	code, err := plugin.NextOrFailure(c.Name(), c.next, ctx, w, r)
	if err != nil {
		return code, err
	}

	return code, nil
}
