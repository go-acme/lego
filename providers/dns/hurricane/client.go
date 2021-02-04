package hurricane

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
)

func (d *DNSProvider) updateTxtRecord(domain string, txt string) error {
	challengeRecord := fmt.Sprintf("_acme-challenge.%s", domain)
	log.Printf("hurricane: Updating record %s", challengeRecord)
	response, err := d.config.HTTPClient.PostForm(
		"https://dyn.dns.he.net/nic/update",
		url.Values{
			"password": {d.config.Token},
			"hostname": {challengeRecord},
			"txt":      {txt},
		})
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	body := strings.TrimSpace(string(bodyBytes))
	switch body {
	case "good":
		return nil
	case "nochg":
		log.Printf("hurricane: nochg: unchanged content written to TXT record %s", challengeRecord)
		return nil
	case "abuse":
		return fmt.Errorf("hurricane: abuse: Blocked hostname for abuse: %s", challengeRecord)
	case "badagent":
		return fmt.Errorf("hurricane: badagent: User agent not sent or HTTP method not recognized; open an issue on go-acme/lego on Github")
	case "badauth":
		return fmt.Errorf("hurricane: badauth: Wrong authentication token provided for TXT record %s", challengeRecord)
	case "nohost":
		return fmt.Errorf("hurricane: nohost: The record provided does not exist in this account: %s", challengeRecord)
	case "notfqdn":
		return fmt.Errorf("hurricane: notfqdn: The record provided isn't an FQDN: %s", challengeRecord)
	default:
		// This is basically only server errors.
		return fmt.Errorf("hurricane: Attempt to change TXT record %s returned %s", challengeRecord, body)
	}
}
