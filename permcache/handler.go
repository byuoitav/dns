package permcache

import (
	"context"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	bolt "go.etcd.io/bbolt"
)

var log = clog.NewWithPlugin("permcache")

const (
	_name = "permcache"
)

type Cache struct {
	Next plugin.Handler

	db *bolt.DB
}

func (c *Cache) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// rw saves answer in the cache when the answer comes back
	rw := &responseWriter{
		ResponseWriter: w,
		cache:          c,
		state:          state,
		remoteAddr:     w.RemoteAddr(),
	}

	// check for the answer in the cache
	msg, err := c.get(r)
	if err != nil {
		// not in cache, just get the item like normal
		if len(r.Question) == 1 {
			log.Infof("%v | Forwarding request '%s' (unable to get from cache: %s)", r.Id, r.Question[0].Name, err)
		} else {
			log.Infof("%v | Forwarding request (unable to get from cache: %s)", r.Id, err)
		}

		rw.writeToClient = true
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, rw, r)
	}

	log.Infof("%v | Returning answer for '%s' from cache and fetching real answer in background", r.Id, r.Question[0].Name)

	// write the cached answer to the client
	_ = w.WriteMsg(msg)

	// fetch the real answer in the background
	// so that the next query will get the correct answer
	go func() {
		rw.writeToClient = false
		if _, err := plugin.NextOrFailure(c.Name(), c.Next, ctx, rw, r); err != nil {
			log.Warningf("%v | unable to fetch real answer: %s", r.Id, err)
		}
	}()

	return dns.RcodeSuccess, nil
}

func (c *Cache) Name() string { return _name }
