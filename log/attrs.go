package log

import (
	"log/slog"
	"strings"
)

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func DomainAttr(v string) slog.Attr {
	return slog.String("domain", v)
}

func DomainsAttr(v []string) slog.Attr {
	return slog.String("domains", strings.Join(v, ", "))
}

func CertNameAttr(v string) slog.Attr {
	return slog.String("cert-name", v)
}
