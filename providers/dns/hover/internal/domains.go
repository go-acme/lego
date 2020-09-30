package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	authHeader = "hoverauth"
)

var (
	// https://www.hover.com/api/domains -> DomainList
	parsedBaseURL = mustParse("https://www.hover.com/api")
)

func mustParse(aURL string) *url.URL {
	if parsed, err := url.Parse(aURL); err != nil {
		panic(fmt.Sprintf("url [%s] unparseable: %v", aURL, err))
	} else {
		return parsed
	}
}

// Address holds an address used for admin, billing, or tech contact.  Empirically, it seems at
// least US and Canada formats are squeezed into a US format.  Please PR if you discover additional
// formats.
type Address struct {
	Status           string `json:"status"`     // Status seems to be "active" in all my zones
	OrganizationName string `json:"org_name"`   // Name of Organization
	FirstName        string `json:"first_name"` // First naem seems to be given non-family name, not positional
	LastName         string `json:"last_name"`  // Last Name seems to be family name, not positional
	Address1         string `json:"address1"`
	Address2         string `json:"address2"`
	Address3         string `json:"address3"`
	City             string `json:"city"`
	State            string `json:"state"`   // State seems to be the US state or the Canadian province
	Zip              string `json:"zip"`     // 5-digit US (ie 10001) or 6-char slammed Canadian (V0H1X0 no space)
	Country          string `json:"country"` // 2-leter state code; this seems to match the second (non-country) of a ISO-3166-2 code
	Phone            string `json:"phone"`   // phone format all over the map, but thy seem to write it as a ITU E164, but a "." separating country code and subscriber number
	Facsimile        string `json:"fax"`     // same format as phone
	Email            string `json:"email"`   // rfc2822 format email address such as rfc2822 para 3.4.1
}

// ContactBlock is merely the four contact addresses that Hover uses, but it's easier to work with
// a defined type in static constants during testing
type ContactBlock struct {
	Admin   Address `json:"admin"`
	Billing Address `json:"billing"`
	Tech    Address `json:"tech"`
	Owner   Address `json:"owner"`
}

// Domain structure describes the config for an entire domain within Hover: the dates involved,
// contact addresses, nameservers, etc: it seems to cover everything about the domain in one
// structure, which is convenient when you want to compare data across many domains.
type Domain struct {
	ID             string       `json:"id"`                        // A unique opaque identifier defined by Hover
	DomainName     string       `json:"domain_name"`               // the actual domain name.  ie: "example.com"
	NumEmails      int          `json:"num_emails,omitempty"`      // This appears to be the number of email accounts either permitted or defined for the domain
	RenewalDate    string       `json:"renewal_date,omitempty"`    // This renewal date appears to be the first day of non-service after a purchased year of valid service: the first day offline if you don't renew.  RFC3339/ISO8601 -formatted yyyy-mm-dd.
	DisplayDate    string       `json:"display_date"`              // Display Date seems to be the same as Renewal Date but perhaps can allow for odd display corner-cases such as leap-years, leap-seconds, or timezones oddities.  RFC3339/ISO8601 to granularity of day as well.
	RegisteredDate string       `json:"registered_date,omitempty"` // Date the domain was first registered, which is likely also the first day of service (or partial-day, technically)  RFC3339/ISO8601 to granularity of day as well.
	Active         bool         `json:"active,omitempty"`          // Domain Entries also show which zones are active
	Contacts       ContactBlock `json:"contacts"`
	Entries        []Entry      `json:"entries,omitempty"` // entries in a zone, if expanded
	HoverUser      User         `json:"hover_user,omitempty"`
	Glue           struct{}     `json:"glue,omitempty"` // I'm not sure how Hover records Glue Records here, or whether they're still used.  Please PR a suggested format!
	NameServers    []string     `json:"nameservers,omitempty"`
	Locked         bool         `json:"locked,omitempty"`
	Renewable      bool         `json:"renewable,omitempty"`
	AutoRenew      bool         `json:"auto_renew,omitempty"`
	Status         string       `json:"status,omitempty"`        // Status seems to be "active" in all my zones
	WhoisPrivacy   bool         `json:"whois_privacy,omitempty"` // boolean as lower-case string: keep your real address out of whois?
}

// Entry is a single DNS record, such as a single NS, TXT, A, PTR, AAAA record within a zone.
type Entry struct {
	CanRevert bool   `json:"can_revert"`
	Content   string `json:"content"`    // free-form text of verbatim value to store (ie "192.168.0.1" for A-rec)
	ID        string `json:"id"`         // A unique opaque identifier defined by Hover
	Default   bool   `json:"is_default"` // seems to track the default @ or "*" record
	Name      string `json:"name"`       // entry name, or "*" for default
	TTL       int    `json:"ttl"`        // TimeToLive, seconds
	Type      string `json:"type"`       // record type: A, MX, PTR, TXT, etc
}

// PlaintextAuth is a structure into which the username and password are read from a plaintext
// file.  This is necessary because when this code is written, Hover offers no API, so raw logins
// are mimicked as clients.  This has risks, of course.  The trade-off is that plaintext risk
// versus no functionality means we have no alternative.  For this reason, reading the auth from a
// file on disk means it cannot be offered in-code during integration tests, and similarly, can be
// provided by a configmap or similar during a production deployment.
//
// For versatility, the intent is to accept JSON, YAML, and even XML if it's trivial to do.
type PlaintextAuth struct {
	Username          string `json:"username"`          // username such as 'chickenandpork', exactly as typed in the login form
	PlaintextPassword string `json:"plaintextpassword"` // password, in plaintext, for login, exactly as typed in the login form
}

// The User record in a Domain seems to record additional contact information that augments the
// Billing Contact with the credit card used and some metadata around it.
type User struct {
	Billing struct {
		Description string `json:"description,omitempty"` // This seems to be a description of my card, such as "Visa ending 1234"
		PayMode     string `json:"pay_mode,omitempty"`    // some reference to how payments are processed:  mine all say "apple_pay", and they're in my Apple
		// Pay Wallet, but my account on Hover predates the existence of Apple Wallet, so ... I'm not sure
	} `json:"billing,omitempty"`
	Email          string `json:"email,omitempty"`
	EmailSecondary string `json:"email_secondary,omitempty"`
}

var (
	// HoverAddress is a constant-ish var that I use to ensure that within my domains, the ones
	// I expect to have Hovers contact info (their default) do.  For example, Tech Contacts
	// where I don't want to be that guy (for managed domains, they should be the tech
	// contact).  Of course, if the values in this constant are incorrect, TuCows is the
	// authority, but please PR me a correction to help me maintain accuracy.
	HoverAddress = Address{
		Status:           "active",
		OrganizationName: "Hover, a service of Tucows.com Co",
		FirstName:        "Support",
		LastName:         "Contact",
		Address1:         "96 Mowat Ave.",
		City:             "Toronto",
		State:            "ON",
		Zip:              "M6K 3M1",
		Country:          "CA",
		Phone:            "+1.8667316556",
		Email:            "help@hover.com",
	}
)

// DomainList is a structure mapping the json response to a request for a list of domains.  It
// tends to be a very rich response including an array of full Domain instances.
type DomainList struct {
	Succeeded bool     `json:"succeeded"`
	Domains   []Domain `json:"domains"`
}

// Client is the client context for communicating with Hover DNS API; should only need one of these
// but keeping state isolated to instances rather than global where possible.
type Client struct {
	HTTPClient *http.Client
	log        YALI       // Yet Another Logger Interface, NopLogger to discard
	authCookie string     // intentionally private
	domains    DomainList // intentionally private
	Username   string
	Password   string
}

// APIURL is an attempt to keep the URLs all based from parsedBaseURL, but more symbollically
// generated and less risk of typos.  The gain on this function is dubious, and this may disappear
//
// TODO: consider rolling in c.BaseURL
func APIURL(resource string) string {
	newURL := *parsedBaseURL
	newURL.Path = fmt.Sprintf("%s/%s", newURL.Path, resource)
	return newURL.String()
}

// APIURLDNS extends the consistency objectives of APIURL by bookending a domain unique ID with
// the /domains/ and /dns pre/post wrappers
func APIURLDNS(domainID string) string {
	return APIURL(fmt.Sprintf("domains/%s/dns", domainID))
}

// GetDomainEntries gets the entries for a specific domain -- essentially the zone records
func (c *Client) GetDomainEntries(domain string) error {
	if _, err := c.GetAuth(); err != nil {
		return fmt.Errorf(`Exception "%s" getting auth for [%s]`, err, APIURLDNS(domain))
	}
	resp, err := c.HTTPClient.Get(APIURLDNS(domain))
	if err != nil {
		c.log.Printf(`Exception "%s" hitting [%s]`, err, APIURLDNS(domain))
		return fmt.Errorf(`Exception "%s" hitting [%s]`, err, APIURLDNS(domain))
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		var nd DomainList
		json.Unmarshal([]byte(body), &nd)

		for n, v := range c.domains.Domains {
			if v.DomainName == domain {
				c.log.Printf(`replacing "%+v"`, c.domains.Domains[n])
				for _, d := range nd.Domains {
					if d.DomainName == domain {
						c.domains.Domains[n] = d
					}
				}
				c.log.Printf(`replaced "%+v"`, c.domains.Domains[n])
			}
		}
		return nil
	}
}

// FillDomains fills the list of domains allocated to the usernamr and password to the Domains
// structure.  It will use GetAuth() to perform a login if necessary.
func (c *Client) FillDomains() error {
	if _, err := c.GetAuth(); err == nil {
		resp, err := c.HTTPClient.Get(APIURL("domains"))
		c.log.Printf("Hitting [%s]\n", APIURL("domains"))
		if err != nil {
			c.log.Printf("hoverdnsapi: GET of %s threw: [%+v].  Domains not expected to be filled.", APIURL("domains"), err)
			return fmt.Errorf("hoverdnsapi: GET of %s threw: [%+v].  Domains not expected to be filled", APIURL("domains"), err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			c.log.Printf("hover: Info: getting domains as user=%s pass=%s returned non-200: Status: %+v StatusCode: %+v\n", c.Username, c.Password, resp.Status, resp.StatusCode)
			return fmt.Errorf("hoverdnsapi: GET of %s as user=%s returned non-200 error: Status: %+v StatusCode: %+v", APIURL("domains"), c.Username, resp.Status, resp.StatusCode)
		} else {
			json.NewDecoder(resp.Body).Decode(&c.domains)
			//c.log.Printf("hover: getting returned: [%+v]\n", c.domains)
		}
	} else {
		c.log.Printf("Auth for user=%s at %s failed\n", c.Username, APIURL("domains"))
		return fmt.Errorf("hoverdnsapi: Auth for GET of %s as user=%s failed", APIURL("domains"), c.Username)
	}
	return nil
}

// ExistingTXTRecords checks whether the given TXT record exists; err != nil if not found
func (c *Client) ExistingTXTRecords(fqdn string) error {
	return fmt.Errorf("hover: (%s) we actually got here: %s", fqdn, c.authCookie)
}

// GetAuth returns the authentication key for the username and password, performing a login if the
// key is not already known from a previous login.
func (c *Client) GetAuth() (string, error) {
	if auth, ok := c.GetCookie(authHeader); ok {
		return auth, nil
	}

	c.log.Printf("Getting fresh authCookie for user=%s at %s\n", c.Username, APIURL("login"))
	req, _ := http.NewRequest("POST", APIURL("login"), strings.NewReader(url.Values{
		"username": {c.Username},
		"password": {c.Password},
	}.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.log.Printf("Error while executing POST: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	c.log.Println(string(body))
	if auth, ok := c.GetCookie(authHeader); ok {
		c.log.Printf("Auth found for user=%s at %s\n", c.Username, APIURL("login"))
		return auth, nil
	}
	return "", fmt.Errorf("hover: No auth in response: %+v -> %s", c.HTTPClient, body)
}

// GetCookie searches existing cookies from a login to Hover's API to find the given cookie.
func (c *Client) GetCookie(key string) (value string, ok bool) {
	if c.HTTPClient == nil {
		return "", false
	}
	if c.HTTPClient.Jar == nil {
		return "", false
	}
	if 1 > len(c.HTTPClient.Jar.Cookies(parsedBaseURL)) {
		c.log.Printf("no cookies for %s", parsedBaseURL)
		return "", false
	}
	c.log.Printf("breaking apart cookies for %+v\n", parsedBaseURL)
	for _, v := range c.HTTPClient.Jar.Cookies(parsedBaseURL) {
		c.log.Printf("k/v: %s/%s\n", v.Name, v.Value)
		if v.Name == key {
			c.log.Printf("returning found: k/v: %s/%s\n", v.Name, v.Value)
			return v.Value, true
		}
	}

	c.log.Printf("Failed to find value for key[%s]\n", key)
	return "", false
}

// GetDomainByName searches iteratively and returns the Domain record that has the given name
func (c *Client) GetDomainByName(domainname string) (*Domain, bool) {
	for _, v := range c.domains.Domains {
		if v.DomainName == domainname {
			return &v, true
		} else {
			c.log.Printf(`Domain "%s" is not objective "%s"\n`, v.DomainName, domainname)
		}
	}

	return nil, false
}

// Upsert inserts or updates a TXT record using the specified parameters
func (c *Client) Upsert(fqdn, domain, value string, ttl uint) error {

	actions := []Action{}
	if err := c.ExistingTXTRecords(fqdn); err == nil {
		actions = append(actions, Action{action: Update, domain: domain, fqdn: fqdn, value: value, ttl: ttl})
	} else {
		actions = append(actions, Action{action: Add, domain: domain, fqdn: fqdn, value: value, ttl: ttl})
	}

	if err := c.DoActions(actions...); err != nil {
		return fmt.Errorf("hover: failed to add record(s) for %s: %w", domain, err)
	}

	return nil
}

// Delete merely enqueues a delete action for DoActions to process
func (c *Client) Delete(fqdn, domain string) error {
	c.log.Printf(`deleting fqdn "%s" from domain "%s"`, fqdn, domain)
	if err := c.DoActions(Action{action: Delete, fqdn: fqdn, domain: domain}); err != nil {
		return fmt.Errorf("hover: failed to delete record for %s: %w", domain, err)
	}

	return nil
}

// HTTPDelete actually does an HTTP call with the DELETE method.  BOG-standard Go only offers GET
// and POST.
//
// TODO: move to a separate file as a layer onto net/http
func (c *Client) HTTPDelete(url string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("HTTPDelete: creating new request: %w", err)
	}

	//req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))

	resp, err := c.HTTPClient.Do(req)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return fmt.Errorf("HTTPDelete: executing delete request: %w", err)
	}
	return nil
}

// HTTPUpdate actually does an HTTP call with the PUT method.  BOG-standard Go only offers GET
// and POST.
//
// # update an existing DNS record: client.call("put", "dns/dns1234567", {"content": "127.0.0.1"})
//
// TODO: I just couldn't get this to work properly: every attempt was 422: Unactionable.  Replaced by delete/add

// GetEntryByFQDN attempts to find a single Entry in the Domain, returning a non-nil result if
// found.  "ok" is manipulated so that an "if" can be used to check whether it was found without
// having to rely on sentinel or implicit values of the returned (ie a nil Entry might not always
// mean "not found", but it does today)
func (d Domain) GetEntryByFQDN(fqdn string) (e *Entry, ok bool) {
	if !strings.HasSuffix(fqdn, "."+d.DomainName) {
		return nil, false
	}

	hostname := fqdn[0:len(fqdn)-len(d.DomainName)-1] + ""
	fmt.Printf("searching for [%s] (ie [%s]) in [%s]\n", hostname, fmt.Sprintf("%s.%s", hostname, d.DomainName), d.DomainName)

	for _, e := range d.Entries {
		if e.Name == hostname {
			return &e, true
		}
	}
	return nil, false
}

// NewClient Creates a Hover client using plaintext passwords against plain username.
// Consider the risk of where the text is stored.
func NewClient(username, password, filename string, timeout time.Duration, opt ...interface{}) *Client {
	j, _ := cookiejar.New(nil)
	var defaultLogger YALI = golog.New(os.Stderr, "", golog.LstdFlags)

	for _, vv := range opt {
		switch v := vv.(type) {
		case YALI:
			defaultLogger = v
		}
	}

	if filename != "" {
		if observed, err := ReadConfigFile(filename); err == nil {
			username = observed.Username
			password = observed.PlaintextPassword
		}
	}
	fmt.Printf("logging in: u: %+v p: %+v f:%+v\n", username, password, filename)

	return &Client{
		HTTPClient: &http.Client{
			Jar:     j,
			Timeout: timeout,
		},
		//BaseURL:    "https://www.hover.com/api/login",
		//Cookie: blank
		//Domains: make(map[string]string, 2),
		Username: username,
		Password: password,
		log:      defaultLogger,
	}
}
