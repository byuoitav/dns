package permcache

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
	bolt "go.etcd.io/bbolt"
)

var (
	errNotInCache = errors.New("not in cache")

	_bucket = []byte(_name)
)

func key(m *dns.Msg) []byte {
	if m.Truncated {
		return nil
	}

	if len(m.Question) != 1 {
		return nil
	}

	buf := []byte{
		byte(m.Question[0].Qtype >> 8),
		byte(m.Question[0].Qtype),
	}

	buf = append([]byte(m.Question[0].Name))
	return buf
}

func msgToVal(m *dns.Msg) ([]byte, error) {
	val := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Rcode:              m.Rcode,
			AuthenticatedData:  m.AuthenticatedData,
			RecursionAvailable: m.RecursionAvailable,
		},
		Answer: m.Answer,
		Ns:     m.Ns,
	}

	// don't copy OPT records
	for _, rr := range m.Extra {
		if rr.Header().Rrtype == dns.TypeOPT {
			continue
		}

		val.Extra = append(val.Extra, rr)
	}

	return val.Pack()
}

// parses b as a response to r
func valToReplyMsg(b []byte, r *dns.Msg) (*dns.Msg, error) {
	resp := &dns.Msg{}
	if err := resp.Unpack(b); err != nil {
		return nil, err
	}

	resp.SetReply(r)

	// update the ttl to 300 (5 minutes)
	ttl := uint32(300)

	for i := range resp.Answer {
		resp.Answer[i].Header().Ttl = ttl
	}

	for i := range resp.Ns {
		resp.Ns[i].Header().Ttl = ttl
	}

	for i := range resp.Extra {
		resp.Extra[i].Header().Ttl = ttl
	}

	return resp, nil
}

func (c *Cache) get(req *dns.Msg) (*dns.Msg, error) {
	key := key(req)
	if len(key) == 0 {
		return nil, errNotInCache
	}

	var val []byte
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(_bucket)
		if b == nil {
			panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
		}

		val = b.Get(key)
		if val == nil {
			return errNotInCache
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return valToReplyMsg(val, req)
}

func (c *Cache) set(req *dns.Msg, resp *dns.Msg) error {
	key := key(req)
	if len(key) == 0 {
		return nil
	}

	val, err := msgToVal(resp)
	if err != nil {
		return err
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(_bucket)
		if b == nil {
			panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
		}

		return b.Put(key, val)
	})
}

func (c *Cache) delete(m *dns.Msg) error {
	key := key(m)
	if len(key) == 0 {
		return nil
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(_bucket)
		if b == nil {
			panic(fmt.Sprintf("bucket '%s' does not exist", _bucket))
		}

		return b.Delete(key)
	})
}
