package permcache

import (
	"net"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type responseWriter struct {
	dns.ResponseWriter
	cache         *Cache
	state         request.Request
	remoteAddr    net.Addr
	writeToClient bool
}

func (w *responseWriter) RemoteAddr() net.Addr {
	if w.remoteAddr != nil {
		return w.remoteAddr
	}

	return w.ResponseWriter.RemoteAddr()
}

func (w *responseWriter) WriteMsg(m *dns.Msg) error {
	if err := w.cache.set(m); err != nil {
		log.Warningf("unable to insert into cache: %s", err)
	}

	if !w.writeToClient {
		return nil
	}

	return w.ResponseWriter.WriteMsg(m)
}
