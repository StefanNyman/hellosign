// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import "fmt"

type EmbeddedAPI struct {
	*hellosign
}

func NewEmbeddedAPI(apiKey string) *EmbeddedAPI {
	return &EmbeddedAPI{newHellosign(apiKey)}
}

type EmbeddedURL struct {
	SignURL   string `json:"sign_url"`
	ExpiresAt uint64 `json:"expires_at"`
}

func (c *EmbeddedAPI) GetSignURL(signatureID string) (*EmbeddedURL, error) {
	resp, err := c.get(fmt.Sprintf("embedded/sign_url/%s", signatureID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	url := &EmbeddedURL{}
	if err := c.parseResponse(resp, url); err != nil {
		return nil, err
	}
	return url, nil
}

func (c *EmbeddedAPI) GetTemplateEditURL(templateID string) (*EmbeddedURL, error) {
	resp, err := c.get(fmt.Sprintf("embedded/edit_url/%s", templateID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	url := &EmbeddedURL{}
	if err := c.parseResponse(resp, url); err != nil {
		return nil, err
	}
	return url, nil
}
