// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/ajg/form"
)

const (
	baseURL                  string = "https://api.hellosign.com/v3"
	contentType                     = "content-type"
	xRatelimitLimit                 = "x-Ratelimit-Limit"
	xRatelimitLimitRemaining        = "x-Ratelimit-Limit-Remaining"
	xRateLimitReset                 = "x-Ratelimit-Reset"
)

type ListInfo struct {
	Page       uint64 `json:"page"`
	NumPages   uint64 `json:"num_pages"`
	NumResults uint64 `json:"num_results"`
	PageSize   uint64 `json:"page_size"`
}

type ListParms struct {
	AccountId *string `form:"account_id,omitempty"`
	Page      *uint64 `form:"page,omitempty"`
	PageSize  *uint64 `form:"page_size,omitempty"`
	Query     *string `form:"query,omitempty"`
}

type FormField struct {
	ApiID    string `json:"api_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	X        uint64 `json:"x"`
	Y        uint64 `json:"y"`
	Width    uint64 `json:"width"`
	Height   uint64 `json:"height"`
	Required bool   `json:"required"`
}

type APIErr struct {
	Code    int // HTTP response code
	Message string
	Name    string
}

type APIWarn struct {
	Code     int // HTTP response code
	Warnings []struct {
		Message string
		Name    string
	}
}

func (a APIErr) Error() string {
	return fmt.Sprintf("%s: %s", a.Name, a.Message)
}

func (a APIWarn) Error() string {
	outMsg := ""
	for _, w := range a.Warnings {
		outMsg += fmt.Sprintf("%s: %s\n", w.Name, w.Message)
	}
	return outMsg
}

type hellosign struct {
	apiKey             string
	RateLimit          uint64 // Number of requests allowed per hour
	RateLimitRemaining uint64 // Remaining number of requests this hour
	RateLimitReset     uint64 // When the limit will be reset. In seconds from epoch
	LastStatusCode     int
}

// Initializes a new Hellosign API client.
func newHellosign(apiKey string) *hellosign {
	return &hellosign{
		apiKey: apiKey,
	}
}

func (c *hellosign) perform(req *http.Request) (*http.Response, error) {
	req.Header.Add("accept", "application/json")
	req.SetBasicAuth(c.apiKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	c.LastStatusCode = resp.StatusCode
	if resp.StatusCode >= 400 {
		return nil, c.parseResponseError(resp)
	}
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
	return resp, err
}

func (c *hellosign) parseResponseError(resp *http.Response) error {
	e := &struct {
		Err struct {
			Msg  *string `json:"error_msg"`
			Name *string `json:"error_name"`
		} `json:"error"`
	}{}
	w := &struct {
		Warnings []struct {
			Msg  *string `json:"warning_msg"`
			Name *string `json:"warning_name"`
		} `json:"warnings"`
	}{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, e)
	if err != nil {
		return err
	}
	if e.Err.Name != nil {
		return APIErr{Code: resp.StatusCode, Message: *e.Err.Msg, Name: *e.Err.Name}
	}
	err = json.Unmarshal(b, w)
	if err != nil {
		return err
	}
	if len(w.Warnings) == 0 {
		return errors.New("Could not parse response error or warning")
	}
	retErr := APIWarn{}
	warns := []struct {
		Name    string
		Message string
	}{}
	for _, w := range w.Warnings {
		warns = append(warns, struct {
			Name    string
			Message string
		}{
			Name:    *w.Name,
			Message: *w.Msg,
		})
	}
	return retErr
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

func GetEptURL(ept string) string {
	return fmt.Sprintf("%s/%s", baseURL, ept)
}

func (c *hellosign) getEptUrl(ept string) string {
	return GetEptURL(ept)
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

func (c *hellosign) getFiles(ept, fileType string, getURL bool) ([]byte, *FileURL, error) {
	if fileType != "" && fileType != "pdf" && fileType != "zip" {
		return []byte{}, nil, errors.New("Invalid file type specified, pdf or zip")
	}
	parms, err := form.EncodeToString(&struct {
		FileType string `form:"file_type,omitempty"`
		GetUrl   bool   `form:"get_url,omitempty"`
	}{
		FileType: fileType,
		GetUrl:   getURL,
	})
	if err != nil {
		return []byte{}, nil, err
	}
	resp, err := c.get(ept, &parms)
	if err != nil {
		return []byte{}, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []byte{}, nil, errors.New(resp.Status)
	}
	if getURL {
		msg := &FileURL{}
		if err := c.parseResponse(resp, msg); err != nil {
			return []byte{}, nil, err
		}
		return []byte{}, msg, nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, nil, err
	}
	return b, nil, nil
}
