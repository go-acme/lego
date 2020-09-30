package internal

import (
	"fmt"
	"net/url"
)

// HoverAct is simply an enum type to typecheck various actions we can perform in a queue
type HoverAct int

const (
	// Error is the zero-value, catch a fencepost error
	Error HoverAct = iota
	// Add a domain record
	Add
	// Delete of a record will require the list of records
	Delete
	// Update will also need the list of records and IDs
	Update
	// Expand is an internal state that will expand a domain to include entries
	Expand
)

// String of course gives a string representation of the Act code
func (h HoverAct) String() string {
	switch h {
	case Error:
		return "--Error--"
	case Add:
		return "Add"
	case Delete:
		return "Delete"
	case Update:
		return "Update"
	case Expand:
		return "--Expand--"
	}

	return "(error) HoverAct const extended without String() equivalent"
}

// Action is a single action (Add, Update, Delete) to complete in a DoActions() call
type Action struct {
	action HoverAct
	fqdn   string
	domain string
	value  string
	ttl    uint
}

func (a Action) String() string {
	return fmt.Sprintf("{action:%s domain:%s fqdn:%s, value:%s ttl:%d}", a.action, a.domain, a.fqdn, a.value, a.ttl)
}

// NewAction returns a newly-created Action as a way to allow external access to create actions
// during test runs using cmd/hoverdns app
func NewAction(action HoverAct, fqdn, domain, value string, ttl uint) Action {
	return Action{action: action, fqdn: fqdn, domain: domain, value: value, ttl: ttl}
}

// DoActions is a way to burn down an accumulated list of actions.  Mostly, this stack will be one
// or two deep, but this offers the chance to go grab a GetAuth() (for authentication cookie) if
// needed, or a detailed DNS list if needed, in a sort of lazy-evaluation logic that avoid these
// actions if not needed.
func (c *Client) DoActions(actions ...Action) (err error) {
	var expansion = map[HoverAct]bool{
		Error:  false,
		Add:    false,
		Delete: true,
		Update: true,
		Expand: false, // of course
	}

	// precondition the steps: prepend expansion where needed
	newActions := make([]Action, 0)
	for n, a := range actions {
		if b, ok := expansion[a.action]; ok && b {
			add := Action{action: Expand, domain: a.domain}
			c.log.Printf("current action (%d), %+v pre-pending %+v", n, a, add)
			newActions = append(newActions, add)
		} else {
			c.log.Printf("current action (%d), %+v no pre-pending", n, a)
		}

		// I just could NOT get update to work reliably, so delete and re-add
		if a.action == Update {
			newActions = append(newActions, Action{action: Delete, fqdn: a.fqdn, domain: a.domain})
			newActions = append(newActions, Action{action: Add, fqdn: a.fqdn, domain: a.domain, value: a.value, ttl: a.ttl})
		} else {
			newActions = append(newActions, a)
		}
	}
	for n, a := range newActions {
		c.log.Printf("resulting actions (%d), %+v", n, a)
	}

	if len(c.domains.Domains) < 1 {
		if err = c.FillDomains(); err != nil {
			return err
		}
	}

	// newActions is a list of:
	// {
	//     action:2 fqdn:_acme-challenge.domain.com domain: domain.com
	//     value:xzLAGicQ1PtUwmXLyCsagNI7O4m_Zsn8mcVREy7QrfY ttl:3600
	// }
	for actnum, a := range newActions {
		if domain, ok := c.GetDomainByName(a.domain); ok {
			fmt.Printf("Action Stack pre (%02d): [%s]\n", actnum, a)
			switch a.action {
			case Error:
				return fmt.Errorf("Error: unset action code: %+v", a)
			case Add:
				if resp, err := c.HTTPClient.PostForm(APIURLDNS(domain.ID), url.Values{
					"name":    {a.fqdn},
					"type":    {"TXT"},
					"content": {a.value},
				}); err != nil {
					fmt.Printf("hover: Info: posting threw: [%+v]\n", err)
				} else {
					resp.Body.Close()
					domain.Entries = make([]Entry, 0) // discard to force refresh on demand
				}
			case Delete:
				if len(domain.Entries) < 1 {
					c.log.Printf(`NOTE: entries for domain "%s" are empty`, domain.DomainName)
				} else if e, ok := domain.GetEntryByFQDN(a.fqdn); !ok {
					c.log.Printf(`NOTE: FQDN "%s" in domain "%s" not found`, a.fqdn, domain.DomainName)
				} else if err := c.HTTPDelete(fmt.Sprintf("%s/%s", APIURLDNS(domain.ID), e.ID)); err != nil {
					c.log.Printf("hover: Info: deleting threw: [%+v]\n", err)
				}

			// As above, I just couldn't get Update to work reliably, so I'm deleting
			// and re-adding as a pair of calls to the Add/Delete defined above in this
			// switch statement.  Therefore, Update case is specifically skipped here.
			//
			//case Update:

			case Expand:
				if len(domain.Entries) < 1 {
					if err := c.GetDomainEntries(a.domain); err != nil {
						return fmt.Errorf("Domain %s not extended: %v", a.domain, err)
					}
				}
			}
			fmt.Printf("Action Stack post (%02d): [%s]\n", actnum, a)
		} else {
			c.log.Printf("domain %s not found", a.domain)
			return fmt.Errorf("Domain %s not found in domains", a.domain)
		}
	}
	return nil
}
