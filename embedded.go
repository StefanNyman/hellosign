// Copyright 2016 Precisely AB.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import "fmt"

// EmbeddedAPI used for embedded signature manipulations.
type EmbeddedAPI struct {
	*hellosign
}

// NewEmbeddedAPI creates a new api client for embedded operations.
func NewEmbeddedAPI(apiKey string) *EmbeddedAPI {
	return &EmbeddedAPI{newHellosign(apiKey)}
}

// EmbeddedURL is an URL with an expiration time.
type EmbeddedURL struct {
	SignURL   string `json:"sign_url"`
	ExpiresAt uint64 `json:"expires_at"`
}

type embeddedURLRaw struct {
	Embedded EmbeddedURL `json:"embedded"`
}

// GetSignURL retrieves an embedded object containing a signature url that can be opened in an iFrame.
func (c *EmbeddedAPI) GetSignURL(signatureID string) (*EmbeddedURL, error) {
	url := &embeddedURLRaw{}
	if err := c.getAndParse(fmt.Sprintf("embedded/sign_url/%s", signatureID), nil, url); err != nil {
		return nil, err
	}
	return &url.Embedded, nil
}

// GetTemplateEditURL retrieves an embedded object containing a template url that can be opened in an iFrame.
// Note that only templates created via the embedded template process are available to be edited with this endpoint.
func (c *EmbeddedAPI) GetTemplateEditURL(templateID string) (*EmbeddedURL, error) {
	url := &embeddedURLRaw{}
	if err := c.getAndParse(fmt.Sprintf("embedded/edit_url/%s", templateID), nil, url); err != nil {
		return nil, err
	}
	return &url.Embedded, nil
}
