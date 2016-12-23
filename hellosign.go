// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/ajg/form"
)

const (
	baseURL                  string = "https://api.hellosign.com/v3"
	contentType                     = "content-type"
	xRatelimitLimit                 = "x-Ratelimit-Limit"
	xRatelimitLimitRemaining        = "x-Ratelimit-Limit-Remaining"
	xRateLimitReset                 = "x-Ratelimit-Reset"
)

type hellosign struct {
	apiKey             string
	baseURL            string
	httpClient         *http.Client
	RateLimit          uint64 // Number of requests allowed per hour
	RateLimitRemaining uint64 // Remaining number of requests this hour
	RateLimitReset     uint64 // When the limit will be reset. In seconds from epoch
	LastStatusCode     int
}

// Initializes a new Hellosign API client.
func newHellosign(apiKey string) *hellosign {
	return &hellosign{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *hellosign) perform(req *http.Request) (*http.Response, error) {
	req.Header.Add("accept", "application/json")
	req.SetBasicAuth(c.apiKey, "")
	d, err := httputil.DumpRequest(req, true)
	if err == nil {
		fmt.Println(string(d))
	}
	resp, err := c.httpClient.Do(req)
	for _, hk := range []string{xRatelimitLimit, xRatelimitLimitRemaining, xRateLimitReset} {
		hv := resp.Header.Get(hk)
		if hv == "" {
			continue
		}
		hvui, pErr := strconv.ParseUint(hv, 10, 64)
		if pErr != nil {
			continue
		}
		switch hk {
		case xRatelimitLimit:
			c.RateLimit = hvui
		case xRatelimitLimitRemaining:
			c.RateLimitRemaining = hvui
		case xRateLimitReset:
			c.RateLimitReset = hvui
		}
	}
	c.LastStatusCode = resp.StatusCode
	return resp, err
}

func (c *hellosign) parseResponse(resp *http.Response, dst interface{}) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		d := json.NewDecoder(resp.Body)
		d.UseNumber()
		return d.Decode(dst)
	}
	return errors.New("Status code invalid")
}

func (c *hellosign) postForm(ept string, o interface{}) (*http.Response, error) {
	v := ""
	if o != nil {
		encoded, err := form.EncodeToString(o)
		if err != nil {
			return nil, err
		}
		v = encoded
	}
	req, err := http.NewRequest(http.MethodPost, c.getEptUrl(ept), strings.NewReader(v))
	if err != nil {
		return nil, err
	}
	req.Header.Add(contentType, "application/x-www-form-urlencoded")
	return c.perform(req)
}

func (c *hellosign) postFormAndParse(ept string, inp, dst interface{}) error {
	resp, err := c.postForm(ept, inp)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp, dst)
}

func (c *hellosign) delete(ept string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, c.getEptUrl(ept), nil)
	if err != nil {
		return nil, err
	}
	return c.perform(req)
}

func (c *hellosign) getEptUrl(ept string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, ept)
}

func (c *hellosign) get(ept string, params *string) (*http.Response, error) {
	url := c.getEptUrl(ept)
	if params != nil && *params == "" {
		url = fmt.Sprintf("%s?%s", url, *params)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.perform(req)
	return resp, err
}

func (c *hellosign) getAndParse(ept string, params *string, dst interface{}) error {
	resp, err := c.get(ept, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp, dst)
}

func (c *hellosign) getFiles(ept, fileType string, getURL bool) (*[]byte, *FileURL, error) {
	if fileType != "" && fileType != "pdf" && fileType != "zip" {
		return nil, nil, errors.New("Invalid file type specified, pdf or zip")
	}
	parms, err := form.EncodeToString(&struct {
		FileType string `form:"file_type,omitempty"`
		GetUrl   bool   `form:"get_url,omitempty"`
	}{
		FileType: fileType,
		GetUrl:   getURL,
	})
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.get(ept, &parms)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New(resp.Status)
	}
	if getURL {
		msg := &FileURL{}
		if err := c.parseResponse(resp, msg); err != nil {
			return nil, nil, err
		}
		return nil, msg, nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return &b, nil, nil
}
