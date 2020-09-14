package permcache

import (
	"bytes"
	"context"
	"fmt"
	"os"

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

	// strip off things that change from request to request
	req := strip(r.Copy())

	log.Infof("NEXT %s", req.Question[0].String())

	code, err := plugin.NextOrFailure(c.Name(), c.Next, ctx, wrapped, r)
	log.Infof("done")
	if err != nil || code != dns.RcodeSuccess {
		log.Errorf("err: %s, code: %s", err, dns.RcodeToString[code])
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

			split := bytes.Split(val, []byte{'\n'})
			for _, b := range split {
				msg := new(dns.Msg)
				if err := msg.Unpack(b); err != nil {
					return fmt.Errorf("unable to unpack message: %w", err)
				}

				msgs = append(msgs, msg)
			}

			return nil
		})

		// nothing was stored in the cache
		// (or there was an error pulling from the cache - we ignore those)
		if len(msgs) == 0 {
			log.Infof("Nothing found in cache, returning error to client")
			return code, err
		}

		for _, msg := range msgs {
			fmt.Printf("writing message:----\n%s----\n\n", msg)
			fmt.Printf("question: %v\n", msg.Question)
			fmt.Printf("ans: %v\n", msg.Answer)
			fmt.Printf("ns: %v\n", msg.Ns)
			fmt.Printf("extra: %v\n", msg.Extra)

			if err := w.WriteMsg(msg); err != nil {
				fmt.Printf("MY error writing message: %s\n", err)
				os.Exit(1)
				return code, err
			}
		}

		os.Exit(1)

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

		buf := new(bytes.Buffer)
		for i, msg := range wrapped.msgs {
			b, err := msg.Pack()
			if err != nil {
				return fmt.Errorf("unable to pack message: %w", err)
			}

			_, _ = buf.Write(b)

			if i < len(wrapped.msgs)-1 {
				_, _ = buf.WriteRune('\n')
			}
		}

		return b.Put([]byte(req.String()), buf.Bytes())
	})
	if err != nil {
		// log error
		fmt.Printf("error: %s\n", err)
	}

	return code, nil
}
