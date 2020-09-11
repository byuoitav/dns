package permcache

import (
	"context"
	"encoding/json"
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

	code, err := plugin.NextOrFailure(c.Name(), c.Next, ctx, wrapped, r)
	if err != nil {
		return code, err
	}

	// strip off things that change from request to request
	req := strip(r.Copy())

	if code != dns.RcodeSuccess {
		// pull answer from db if one exists
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

			if err := json.Unmarshal(val, &msgs); err != nil {
				return err
			}

			return nil
		})
		if len(msgs) == 0 {
			return code, nil
		}

		for _, msg := range msgs {
			if err := w.WriteMsg(msg); err != nil {
				return code, err
			}
		}

		return dns.RcodeSuccess, nil
	}

	// save answer into db
	err = c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(_bucket)
		if b == nil {
			panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
		}

		//fmt.Printf("req:\n-------------------\n%s-------------------\n\n", key)
		//for _, msg := range wrapped.msgs {
		//	fmt.Printf("msg:\n-------------------\n%s-------------------\n\n", msg)
		//}

		wrapped.Lock()
		defer wrapped.Unlock()

		val, err := json.Marshal(wrapped.msgs)
		if err != nil {
			return fmt.Errorf("unable to marshal messages: %w", err)
		}

		return b.Put([]byte(req.String()), val)
	})
	if err != nil {
		// log error
		fmt.Printf("error: %s\n", err)
	}

	return code, nil
}
