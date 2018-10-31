package emca

import (
	"bytes"
	"fmt"
)

// ObtainError is returned when there are specific errors available
// per domain. For example in ObtainCertificate
type ObtainError map[string]error

func (e ObtainError) Error() string {
	buffer := bytes.NewBufferString("acme: Error -> One or more domains had a problem:\n")
	for dom, err := range e {
		buffer.WriteString(fmt.Sprintf("[%s] %s\n", dom, err))
	}
	return buffer.String()
}

type domainError struct {
	Domain string
	Error  error
}
