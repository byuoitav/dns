package permcache

import (
	"bytes"
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	bolt "go.etcd.io/bbolt"
)

var log = clog.NewWithPlugin("permcache")

const (
	_name = "permcache"
)

var (
	_bucket   = []byte(_name)
	_msgSplit = []byte("\r\n")
)

type Cache struct {
	Next plugin.Handler

	db *bolt.DB
}

type Record struct {
	Name    string
	Type    string
	Content string
}

func (c Cache) Name() string { return _name }

func (c Cache) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	wrapped := &responseWriter{
		ResponseWriter: w,
	}

	// strip off things that change from request to request
	req := strip(r.Copy())

	if len(req.Question) == 1 {
		log.Debugf("Resolving %s", req.Question[0].String())
	} else {
		log.Infof("Resolving query with %v questions (???)", len(req.Question))
	}

	code, err := plugin.NextOrFailure(c.Name(), c.Next, ctx, wrapped, r)
	if err != nil {
		log.Warningf("error resolving: %s (Rcode: %s). Attempting to serve from cache", err, dns.RcodeToString[code])

		var msgs []*dns.Msg
		c.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(_bucket)
			if b == nil {
				panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
			}

			val := b.Get([]byte(req.String()))
			if val == nil {
				return nil
			}

			split := bytes.Split(val, _msgSplit)
			for _, b := range split {
				msg := &dns.Msg{}
				if err := msg.Unpack(b); err != nil {
					log.Errorf("unable to unpack message: %w", err)
					continue
				}

				msgs = append(msgs, msg)
			}

			return nil
		})

		// nothing was stored in the cache
		// or there was an error unpacking the message in the cache
		if len(msgs) == 0 {
			log.Infof("No messages found in cache, returning original error to client")
			return code, err
		}

		defer w.Close()

		for _, msg := range msgs {
			if err := w.WriteMsg(msg); err != nil {
				log.Errorf("unable to write message to client: %s", err)
				return code, err
			}
		}

		log.Infof("Successfully returned answer from cache")
		return dns.RcodeSuccess, nil
	}

	// save answer into db
	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(_bucket)
		if b == nil {
			panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
		}

		wrapped.Lock()
		defer wrapped.Unlock()

		buf := &bytes.Buffer{}
		for i, msg := range wrapped.msgs {
			b, err := msg.Pack()
			if err != nil {
				return fmt.Errorf("unable to pack message: %w", err)
			}

			_, _ = buf.Write(b)

			// split up messages for parsing later
			if i < len(wrapped.msgs)-1 {
				_, _ = buf.Write(_msgSplit)
			}
		}

		return b.Put([]byte(req.String()), buf.Bytes())
	})
	if err != nil {
		clog.Errorf("unable to cache answer: %s", err)
	}

	clog.Debugf("Successfully cached and returned answer")
	return code, nil
}
