package certificate

import (
	"bytes"
	"fmt"
)

// obtainError is returned when there are specific errors available
// per domain. For example in Obtain
type obtainError map[string]error

func (e obtainError) Error() string {
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
