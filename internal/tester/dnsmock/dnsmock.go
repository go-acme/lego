package dnsmock

import (
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

const noType uint16 = math.MaxUint16

type Option func(*dns.Server) error

type Builder struct {
	// domain -> op -> type
	routes map[string]map[int]map[uint16]dns.Handler

	stringToType map[string]uint16
}

func NewServer() *Builder {
	stringToType := make(map[string]uint16)
	for typ, str := range dns.TypeToString {
		stringToType[str] = typ
	}

	return &Builder{
		routes:       make(map[string]map[int]map[uint16]dns.Handler),
		stringToType: stringToType,
	}
}

func (b *Builder) Query(pattern string, handler dns.HandlerFunc) *Builder {
	route, err := b.route(pattern, dns.OpcodeQuery, handler)
	if err != nil {
		panic(err.Error())
	}

	return route
}

func (b *Builder) Update(pattern string, handler dns.HandlerFunc) *Builder {
	route, err := b.route(pattern, dns.OpcodeUpdate, handler)
	if err != nil {
		panic(err.Error())
	}

	return route
}

func (b *Builder) route(pattern string, op int, handler dns.HandlerFunc) (*Builder, error) {
	parts := strings.Fields(pattern)

	domain := parts[0]

	_, ok := dns.IsDomainName(domain)
	if !ok {
		return nil, fmt.Errorf("%s: invalid domain: %s", dns.OpcodeToString[op], domain)
	}

	if _, ok := b.routes[domain]; !ok {
		b.routes[domain] = make(map[int]map[uint16]dns.Handler)
	}

	if _, ok := b.routes[domain][op]; !ok {
		b.routes[domain][op] = make(map[uint16]dns.Handler)
	}

	if _, ok := b.routes[domain][op][noType]; ok {
		return nil, fmt.Errorf("%s: a global route already exists for the domain: %s", dns.OpcodeToString[op], domain)
	}

	switch len(parts) {
	case 1:
		if len(b.routes[domain][op]) > 0 {
			return nil, fmt.Errorf("%s: global route and specific routes cannot be mixed for the same domain: %s", dns.OpcodeToString[op], domain)
		}

		b.routes[domain][op][noType] = handler

		return b, nil

	case 2:
		raw := parts[1]

		qType, ok := b.stringToType[raw]
		if !ok {
			return nil, fmt.Errorf("%s: unknown type: %s", dns.OpcodeToString[op], raw)
		}

		if _, ok := b.routes[domain][op][qType]; ok {
			return nil, fmt.Errorf("%s: duplicate route: %s", dns.OpcodeToString[op], pattern)
		}

		b.routes[domain][op][qType] = handler

		return b, nil

	default:
		return nil, fmt.Errorf("%s: invalid pattern: %s", dns.OpcodeToString[op], pattern)
	}
}

func (b *Builder) Build(t *testing.T, options ...Option) net.Addr {
	t.Helper()

	mux := dns.NewServeMux()

	server := &dns.Server{
		Addr:         "127.0.0.1:0",
		Net:          "udp",
		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
		Handler:      mux,
		MsgAcceptFunc: func(dh dns.Header) dns.MsgAcceptAction {
			// bypass defaultMsgAcceptFunc to allow dynamic update (https://github.com/miekg/dns/pull/830)
			return dns.MsgAccept
		},
	}

	for _, option := range options {
		require.NoError(t, option(server))
	}

	for pattern, ops := range b.routes {
		mux.HandleFunc(pattern, func(w dns.ResponseWriter, req *dns.Msg) {
			mTypes, ok := ops[req.Opcode]
			if !ok {
				_ = w.WriteMsg(new(dns.Msg).SetRcode(req, dns.RcodeNotImplemented))

				return
			}

			if h, found := mTypes[noType]; found {
				h.ServeDNS(w, req)

				return
			}

			// For safety but it doesn't happen.
			if len(req.Question) == 0 {
				_ = w.WriteMsg(new(dns.Msg).SetRcode(req, dns.RcodeRefused))

				return
			}

			// For safety but it doesn't happen.
			if req.Question[0].Qclass != dns.ClassINET {
				_ = w.WriteMsg(new(dns.Msg).SetRcode(req, dns.RcodeRefused))

				return
			}

			// Works only for [Query].
			h, ok := mTypes[req.Question[0].Qtype]
			if !ok {
				_ = w.WriteMsg(new(dns.Msg).SetRcode(req, dns.RcodeNotImplemented))

				return
			}

			h.ServeDNS(w, req)
		})
	}

	t.Cleanup(func() {
		_ = server.Shutdown()
	})

	waitLock := sync.Mutex{}
	waitLock.Lock()

	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			t.Log(err)
		}
	}()

	waitLock.Lock()

	return server.PacketConn.LocalAddr()
}
