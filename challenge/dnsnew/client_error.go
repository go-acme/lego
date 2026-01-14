package dnsnew

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// DNSError error related to DNS calls.
type DNSError struct {
	Message string
	NS      string
	MsgIn   *dns.Msg
	MsgOut  *dns.Msg
	Err     error
}

func (d *DNSError) Error() string {
	var details []string
	if d.NS != "" {
		details = append(details, "ns="+d.NS)
	}

	if d.MsgIn != nil && len(d.MsgIn.Question) > 0 {
		details = append(details, fmt.Sprintf("question='%s'", formatQuestions(d.MsgIn.Question)))
	}

	if d.MsgOut != nil {
		if d.MsgIn == nil || len(d.MsgIn.Question) == 0 {
			details = append(details, fmt.Sprintf("question='%s'", formatQuestions(d.MsgOut.Question)))
		}

		details = append(details, "code="+dns.RcodeToString[d.MsgOut.Rcode])
	}

	msg := "DNS error"
	if d.Message != "" {
		msg = d.Message
	}

	if d.Err != nil {
		msg += ": " + d.Err.Error()
	}

	if len(details) > 0 {
		msg += " [" + strings.Join(details, ", ") + "]"
	}

	return msg
}

func (d *DNSError) Unwrap() error {
	return d.Err
}

func formatQuestions(questions []dns.Question) string {
	var parts []string
	for _, question := range questions {
		parts = append(parts, strings.ReplaceAll(strings.TrimPrefix(question.String(), ";"), "\t", " "))
	}

	return strings.Join(parts, ";")
}
