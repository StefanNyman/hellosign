// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/ajg/form"
)

type APIAppAPI struct {
	*hellosign
}

func NewAPIAppAPI(apiKey string) *APIAppAPI {
	return &APIAppAPI{newHellosign(apiKey)}
}

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

func (c *APIAppAPI) Get(clientID string) (*APIApp, error) {
	app := &apiAppRaw{}
	if err := c.getAndParse(fmt.Sprintf("api_app/%s", clientID), nil, app); err != nil {
		return nil, err
	}
	return &app.APIApp, nil
}

type APIAppLst struct {
	ListInfo ListInfo `json:"list_info"`
	APIApps  []APIApp `json:"api_apps"`
}

func (c *APIAppAPI) List(parms ListParms) (*APIAppLst, error) {
	paramString, err := form.EncodeToString(parms)
	if err != nil {
		return nil, err
	}
	lst := &APIAppLst{}
	if err := c.getAndParse("api_app/list", &paramString, lst); err != nil {
		return nil, err
	}
	return lst, nil
}

type APIAppCreateParms struct {
	Name                 string             `form:"name"`
	Domain               string             `form:"domain"`
	CallbackURL          string             `form:"callback_url,omitempty"`
	CustomLogoFile       []byte             `form:"custom_logo_file,omitempty"`
	OAuth                *APIAppCreateOauth `form:"oauth,omitempty"`
	WhiteLabelingOptions string             `form:"white_labeling_options,omitempty`
}

type APIAppCreateOauth struct {
	CallbackURL string   `form:"callback_url,omitempty"`
	Scopes      []string `form:"scopes,omitempty"`
}

func (c *APIAppAPI) Create(parms APIAppCreateParms) (*APIApp, error) {
	app := &apiAppRaw{}
	if err := c.postFormAndParse("api_app", &parms, app); err != nil {
		return nil, err
	}
	return &app.APIApp, nil
}

type APIAppUpdateParms struct {
	Name                 string              `form:"name,omitempty"`
	Domain               string              `form:"domain,omitempty"`
	CallbackURL          string              `form:"callback_url,omitempty"`
	CustomLogoFile       []byte              `form:"custom_logo_file,omitempty"`
	OAuth                []APIAppUpdateOauth `form:"oauth,omitempty"`
	WhiteLabelingOptions string              `form:"white_labeling_options,omitempty"`
}

type APIAppUpdateOauth struct {
	CallbackURL string   `form:"callback_url,omitempty"`
	Scopes      []string `form:"scopes,omitempty"`
}

func (c *APIAppAPI) Update(clientID string, parms APIAppUpdateParms) (*APIApp, error) {
	app := &apiAppRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("api_app/%s", clientID), &parms, app); err != nil {
		return nil, err
	}
	return &app.APIApp, nil
}

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
