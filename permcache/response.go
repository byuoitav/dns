package permcache

import (
	"sync"

	"github.com/miekg/dns"
)

type responseWriter struct {
	dns.ResponseWriter

	sync.Mutex
	msgs []*dns.Msg
}

func (w *responseWriter) WriteMsg(msg *dns.Msg) error {
	w.Lock()
	defer w.Unlock()

	w.msgs = append(w.msgs, strip(msg.Copy()))

	// don't write failures
	if msg.Opcode == dns.RcodeSuccess {
		return w.ResponseWriter.WriteMsg(msg)
	}

	return nil
}

func strip(m *dns.Msg) *dns.Msg {
	m.Id = 0
	m.Extra = nil
	return m
}
