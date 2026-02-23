package dnspersist01

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func displayRecordCreationInstructions(fqdn, value string) {
	fmt.Printf("dnspersist01: Please create a TXT record with the following value:\n")
	fmt.Printf("%s IN TXT %s\n", fqdn, formatTXTValue(value))
	fmt.Printf("dnspersist01: Press 'Enter' once the record is available\n")
}

// formatTXTValue formats a TXT record value for display, splitting it into
// multiple quoted strings if it exceeds 255 octets, as per RFC 1035.
func formatTXTValue(value string) string {
	chunks := splitTXTValue(value)
	if len(chunks) == 1 {
		return fmt.Sprintf("%q", chunks[0])
	}

	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, fmt.Sprintf("%q", chunk))
	}

	return strings.Join(parts, " ")
}

// splitTXTValue splits a TXT value into RFC 1035 <character-string> chunks of
// at most 255 octets so long TXT values can be represented as multiple strings
// in one RR.
func splitTXTValue(value string) []string {
	const maxTXTStringOctets = 255
	if len(value) <= maxTXTStringOctets {
		return []string{value}
	}

	var chunks []string
	for len(value) > maxTXTStringOctets {
		chunks = append(chunks, value[:maxTXTStringOctets])
		value = value[maxTXTStringOctets:]
	}
	if len(value) > 0 {
		chunks = append(chunks, value)
	}

	return chunks
}

func waitForUser() error {
	_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	return nil
}
