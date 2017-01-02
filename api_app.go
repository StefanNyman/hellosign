// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"fmt"

	"net/http"
)

// APIAppAPI used for api app manipulations.
type APIAppAPI struct {
	*hellosign
}

// NewAPIAppAPI creates a new api client for api app operations.
func NewAPIAppAPI(apiKey string) *APIAppAPI {
	return &APIAppAPI{newHellosign(apiKey)}
}

// APIApp Contains information about an API App.
type APIApp struct {
	CallbackURL *string `json:"callback_url"`
	ClientID    string  `json:"client_id"`
	CreatedAt   uint64  `json:"created_at"`
	Domain      string  `json:"domain"`
	IsApproved  bool    `json:"is_approved"`
	Name        string  `json:"name"`
	Oauth       *struct {
		CallbackURL string   `json:"callback_url"`
		Scopes      []string `json:"scopes"`
		Secret      string   `json:"string"`
	} `json:"oauth"`
	OwnerAccount struct {
		AccountID    string `json:"account_id"`
		EmailAddress string `json:"email_address"`
	} `json:"owner_account"`
}

type apiAppRaw struct {
	APIApp APIApp `json:"api_app"`
}

// Get returns a struct with information about an API App.
func (c *APIAppAPI) Get(clientID string) (*APIApp, error) {
	app := &apiAppRaw{}
	err := c.getAndParse(fmt.Sprintf("api_app/%s", clientID), nil, app)
	return &app.APIApp, err
}

// APIAppLst is a list of api apps managed by this account.
type APIAppLst struct {
	ListInfo ListInfo `json:"list_info"`
	APIApps  []APIApp `json:"api_apps"`
}

// List returns a list of API Apps that are accessible by you. If you are on a team with an Admin
// or Developer role, this list will include apps owned by teammates.
func (c *APIAppAPI) List(parms ListParms) (*APIAppLst, error) {
	lst := &APIAppLst{}
	err := c.list("api_app/list", parms, lst)
	return lst, err
}

// APIAppCreateParms parameters for creating an api app.
type APIAppCreateParms struct {
	Name                 string             `form:"name"`
	Domain               string             `form:"domain"`
	CallbackURL          string             `form:"callback_url,omitempty"`
	CustomLogoFile       []byte             `form:"custom_logo_file,omitempty"`
	OAuth                *APIAppCreateOauth `form:"oauth,omitempty"`
	WhiteLabelingOptions string             `form:"white_labeling_options,omitempty"`
}

// APIAppCreateOauth OAuth params that can be provided when creating an app.
type APIAppCreateOauth struct {
	CallbackURL string   `form:"callback_url,omitempty"`
	Scopes      []string `form:"scopes,omitempty"`
}

// Create Creates a new API App.
func (c *APIAppAPI) Create(parms APIAppCreateParms) (*APIApp, error) {
	app := &apiAppRaw{}
	if err := c.postFormAndParse("api_app", &parms, app); err != nil {
		return nil, err
	}
	return &app.APIApp, nil
}

// APIAppUpdateParms parameters for updating an api app.
type APIAppUpdateParms struct {
	Name                 string              `form:"name,omitempty"`
	Domain               string              `form:"domain,omitempty"`
	CallbackURL          string              `form:"callback_url,omitempty"`
	CustomLogoFile       []byte              `form:"custom_logo_file,omitempty"`
	OAuth                []APIAppUpdateOauth `form:"oauth,omitempty"`
	WhiteLabelingOptions string              `form:"white_labeling_options,omitempty"`
}

// APIAppUpdateOauth OAuth params that can be provided when updating an app.
type APIAppUpdateOauth struct {
	CallbackURL string   `form:"callback_url,omitempty"`
	Scopes      []string `form:"scopes,omitempty"`
}

// Update Updates an existing API App. Can only be invoked for apps you own. Only the fields you
// provide will be updated. If you wish to clear an existing optional field, provide an empty string.
func (c *APIAppAPI) Update(clientID string, parms APIAppUpdateParms) (*APIApp, error) {
	app := &apiAppRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("api_app/%s", clientID), &parms, app); err != nil {
		return nil, err
	}
	return &app.APIApp, nil
}

// Delete deletes an API App. Can only be invoked for apps you own.
func (c *APIAppAPI) Delete(clientID string) (bool, error) {
	resp, err := c.delete(fmt.Sprintf("api_app/%s", clientID))
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return false, errors.New(resp.Status)
	}
	return true, nil
}
