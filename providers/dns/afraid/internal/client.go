package internal

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-acme/lego/v4/log"
)

const (
	logonURL        = "https://freedns.afraid.org/zc.php?step=2"
	domainsURL      = "https://freedns.afraid.org/domain/"
	subdomainURL    = "https://freedns.afraid.org/subdomain/"
	editRecordURL   = "https://freedns.afraid.org/subdomain/save.php?step=2"
	deleteRecordURL = "https://freedns.afraid.org/subdomain/delete2.php?submit=delete selected"
)

type Client struct {
	Login           string
	Password        string
	IsAuthenticated bool
	HTTPClient      *http.Client
	LogonURL        *url.URL
	DomainsURL      *url.URL
	SubdomainsURL   *url.URL
	EditRecordURL   *url.URL
	DeleteRecordURL *url.URL
	Domains         map[string]uint64
	authOnce        sync.Once
	domainsOnce     sync.Once
}

func NewClient(login string, password string) (*Client, error) {
	logonURL, err := url.Parse(logonURL)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	domainsURL, err := url.Parse(domainsURL)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	subdomainsURL, err := url.Parse(subdomainURL)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	editRecordURL, err := url.Parse(editRecordURL)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	deleteRecordURL, err := url.Parse(deleteRecordURL)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("afraid: %w", err)
	}

	httpClient := &http.Client{}
	httpClient.Jar = jar

	client := &Client{
		Login:           login,
		Password:        password,
		IsAuthenticated: false,
		LogonURL:        logonURL,
		DomainsURL:      domainsURL,
		SubdomainsURL:   subdomainsURL,
		EditRecordURL:   editRecordURL,
		DeleteRecordURL: deleteRecordURL,
		HTTPClient:      httpClient,
		Domains:         make(map[string]uint64),
	}
	return client, nil
}

func (c *Client) SetHTTPTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

func (c *Client) CreateTxtRecord(fqdn string, value string) error {
	err := c.authenticate()
	if err != nil {
		return err
	}

	err = c.loadDomains()
	if err != nil {
		return err
	}

	domainID, recordName, recordID, err := c.getDomainAndRecord(fqdn)
	if err != nil {
		return err
	}

	values := make(url.Values)
	values.Set("domain_id", strconv.FormatUint(domainID, 10))
	values.Set("subdomain", recordName)
	values.Set("address", fmt.Sprintf("\"%s\"", value))
	values.Set("type", "TXT")
	values.Set("send", "Save!")
	if recordID > 0 {
		values.Set("data_id", strconv.FormatUint(recordID, 10))
	}

	response, err := c.HTTPClient.PostForm(c.EditRecordURL.String(), values)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("cannot create TXT record %s: %w", fqdn, err)
	}

	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	recordID = c.findRecordID(html, fqdn)
	if recordID == 0 {
		return fmt.Errorf("new record %s not found", fqdn)
	}

	return nil
}

func (c *Client) DeleteTxtRecord(fqdn string) error {
	err := c.authenticate()
	if err != nil {
		return err
	}

	err = c.loadDomains()
	if err != nil {
		return err
	}

	_, _, recordID, err := c.getDomainAndRecord(fqdn)
	if err != nil {
		return err
	}
	if recordID == 0 {
		return fmt.Errorf("record %s not found", fqdn)
	}

	deleteURL := *c.DeleteRecordURL
	urlQuery := deleteURL.Query()
	urlQuery.Add("data_id[]", strconv.FormatUint(recordID, 10))
	deleteURL.RawQuery = urlQuery.Encode()

	response, err := c.HTTPClient.Get(deleteURL.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("cannot delete TXT record %s: %w", fqdn, err)
	}

	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	recordID = c.findRecordID(html, fqdn)
	if recordID != 0 {
		return fmt.Errorf("record %s exists after delete query", fqdn)
	}

	return nil
}

func (c *Client) authenticate() error {
	var err error
	c.authOnce.Do(func() {
		values := make(url.Values)
		values.Set("username", c.Login)
		values.Set("password", c.Password)
		values.Set("submit", "Login")
		values.Set("action", "auth")

		var response *http.Response
		response, err = c.HTTPClient.PostForm(c.LogonURL.String(), values)
		if err != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode >= 400 {
			err = fmt.Errorf("authentication error %s", response.Status)
			return
		}

		cookieFound := false
		for _, cookie := range c.HTTPClient.Jar.Cookies(c.LogonURL) {
			if cookie.Name == "dns_cookie" {
				cookieFound = true
				break
			}
		}
		if !cookieFound {
			err = fmt.Errorf("session cookie not found")
			return
		}

		var html *goquery.Document
		html, err = goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return
		}
		title := html.Find("title")
		log.Infof("afraid: %s", strings.TrimSpace(title.Text()))

		c.IsAuthenticated = true
	})
	if err != nil {
		return err
	}
	if !c.IsAuthenticated {
		return errors.New("not authorized")
	}
	return nil
}

func (c *Client) loadDomains() error {
	var err error
	c.domainsOnce.Do(func() {
		var response *http.Response
		response, err = c.HTTPClient.Get(c.DomainsURL.String())
		if err != nil {
			return
		}
		defer response.Body.Close()

		var html *goquery.Document
		html, err = goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return
		}
		links := html.Find("a[href^=\"/subdomain/?limit=\"]")

		domainIDRegex := regexp.MustCompile(`/subdomain/\?limit=([0-9]+)`)

		links.Each(func(i int, link *goquery.Selection) {
			href, exist := links.Attr("href")
			if !exist || !domainIDRegex.MatchString(href) {
				return
			}
			domain := link.Closest("td").Find("b").First().Text()
			domainIDString := domainIDRegex.FindStringSubmatch(href)[1]
			domainID, errInt := strconv.ParseUint(domainIDString, 10, 64)
			if errInt != nil {
				log.Fatalf("afraid: domain %s has invalid id %s (%w)", domain, domainIDString, err)
				return
			}
			c.Domains[domain] = domainID
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) getRecordID(domainID uint64, fqdn string) (uint64, error) {
	subdomainsURL := *c.SubdomainsURL
	urlQuery := subdomainsURL.Query()
	urlQuery.Add("limit", strconv.FormatUint(domainID, 10))
	subdomainsURL.RawQuery = urlQuery.Encode()

	response, err := c.HTTPClient.Get(subdomainsURL.String())
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return 0, err
	}

	recordID := c.findRecordID(html, fqdn)

	return recordID, nil
}

func (c *Client) findRecordID(html *goquery.Document, fqdn string) uint64 {
	links := html.Find("a[href^=\"edit.php?data_id=\"]")

	recordIDRegex := regexp.MustCompile(`edit.php\?data_id=([0-9]+)`)

	recordID := uint64(0)
	links.EachWithBreak(func(i int, link *goquery.Selection) bool {
		href, exist := link.Attr("href")
		if !exist || !recordIDRegex.MatchString(href) {
			return true
		}
		linkValue := link.Text()
		if linkValue+"." == fqdn {
			recordIDString := recordIDRegex.FindStringSubmatch(href)[1]
			id, errInt := strconv.ParseUint(recordIDString, 10, 64)
			if errInt != nil {
				log.Fatalf("afraid: record %s has invalid id %s (%w)", linkValue, recordIDRegex, errInt)
				return true
			}
			recordID = id
			return false
		}
		return true
	})

	return recordID
}

func (c *Client) getDomainAndRecord(fqdn string) (uint64, string, uint64, error) {
	domainID := uint64(0)
	var recordName string
	for k, v := range c.Domains {
		if strings.HasSuffix(fqdn, k+".") {
			domainID = v
			recordName = strings.TrimSuffix(fqdn, "."+k+".")
			break
		}
	}

	if domainID == 0 {
		return 0, "", 0, fmt.Errorf("domain for %s not found", fqdn)
	}

	recordID, err := c.getRecordID(domainID, fqdn)
	if err != nil {
		return 0, "", 0, err
	}

	return domainID, recordName, recordID, nil
}
