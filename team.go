// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"net/http"
)

type TeamAPI struct {
	*hellosign
}

func NewTeamAPI(apiKey string) *TeamAPI {
	return &TeamAPI{newHellosign(apiKey)}
}

type Team struct {
	Name     string `json:"name"`
	Accounts []struct {
		AccountID    string `json:"account_id"`
		EmailAddress string `json:"email_address"`
		RoleCode     string `json:"role_code"`
	} `json:"accounts"`
	InvitedAccounts []struct {
		AccountID    string `json:"account_id"`
		EmailAddress string `json:"email_address"`
		RoleCode     string `json:"role_code"`
	} `json:"invited_accounts"`
}

type teamRaw struct {
	Team Team `json:"team"`
}

func (c *TeamAPI) Get() (*Team, error) {
	team := &teamRaw{}
	if err := c.getAndParse("team", nil, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}

func (c *TeamAPI) Create(name string) (*Team, error) {
	team := &teamRaw{}
	if err := c.postFormAndParse("team/create", &struct {
		Name string `form:"name"`
	}{
		Name: name,
	}, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}

func (c *TeamAPI) Update(name string) (*Team, error) {
	team := &teamRaw{}
	if err := c.postFormAndParse("team", &struct {
		Name string `form:"name"`
	}{
		Name: name,
	}, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}

func (c *TeamAPI) Delete() (bool, error) {
	resp, err := c.postForm("team/destroy", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	return true, nil
}

func (c *TeamAPI) AddUser(accountID, emailAddress *string) (*Team, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	team := &teamRaw{}
	if err := c.postFormAndParse("team/add_member", &struct {
		AccountID    *string `form:"account_id,omitempty"`
		EmailAddress *string `form:"email_address,omitempty"`
	}{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}

func (c *TeamAPI) RemoveUser(accountID, emailAddress *string) (*Team, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	team := &teamRaw{}
	if err := c.postFormAndParse("team/remove_member", &struct {
		AccountID    *string `form:"account_id,omitempty"`
		EmailAddress *string `form:"email_address,omitempty"`
	}{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}
