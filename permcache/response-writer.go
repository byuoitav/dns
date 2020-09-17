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
	if len(w.state.Req.Question) == 1 {
		log.Infof("%v | Saving answer for '%s' in cache", w.state.Req.Id, w.state.Req.Question[0].Name)
	} else {
		log.Infof("%v | Saving answer in cache (%#q)", w.state.Req.Id, w.state.Req.String())
	}

	if err := w.cache.set(w.state.Req, m); err != nil {
		log.Warningf("%v | unable to insert into cache: %s", w.state.Req.Id, err)
	}

	if !w.writeToClient {
		return nil
	}

	return w.ResponseWriter.WriteMsg(m)
}

func (w *responseWriter) Write(buf []byte) (int, error) {
	log.Warning("can't handle Write() call; not caching reply")
	return w.ResponseWriter.Write(buf)
}
