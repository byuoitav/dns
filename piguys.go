/*
package piguys

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	bolt "go.etcd.io/bbolt"
)

type PiGuys struct{}

const (
	name    = "piguys"
	_bucket = "dns"
)

func (p PiGuys) Name() string { return name }

func (p PiGuys) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	db, err := bolt.Open(os.Getenv("CACHE_DATABASE_LOCATION"), 0600, nil)
	if err != nil {
		return dns.RcodeServerFailure, fmt.Errorf("unable to open cache: %w", err)
	}

	a := &dns.Msg{}
	a.SetReply(r)
	a.Authoritative = true

	ip := state.IP()
	if ip == "" {
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(_bucket))
			if b == nil {
				return fmt.Errorf("bucket does not exist")
			}

			hostname, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("unable to get hostname: %w", err)
			}

			bytes := b.Get([]byte(hostname))
			if bytes == nil {
				return fmt.Errorf("dns response not in cache")
			}

			if err := json.Unmarshal(bytes, &a); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return dns.RcodeServerFailure, fmt.Errorf("unable to find cached dns response: %w", err)
		}

		w.WriteMsg(a)

		return dns.RcodeSuccess, nil
	}
	var rr dns.RR

	switch state.Family() {
	case 1:
		rr = &dns.A{}
		rr.(*dns.A).A = net.ParseIP(ip).To4()
	case 2:
		rr = &dns.AAAA{}
		rr.(*dns.AAAA).AAAA = net.ParseIP(ip)
	}

	a.Extra = []dns.RR{rr}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(_bucket))
		if err != nil {
			return fmt.Errorf("error creating bucket: %w", err)
		}

		bytes, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("unable to marshal dns response: %w", err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("unable to get hostname: %w", err)
		}

		if err = b.Put([]byte(hostname), bytes); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return dns.RcodeServerFailure, fmt.Errorf("unable to cache dns response: %w", err)
	}

	w.WriteMsg(a)

	return dns.RcodeSuccess, nil
}
*/
