// Copyright 2016 Precisely AB.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"net/http"
)

// TeamAPI used for team manipulations
type TeamAPI struct {
	*hellosign
}

// NewTeamAPI creates a new api client for team endpoints.
func NewTeamAPI(apiKey string) *TeamAPI {
	return &TeamAPI{newHellosign(apiKey)}
}

// Team contains information about your team and its members.
type Team struct {
	Name            string    `json:"name"`
	Accounts        []TeamAcc `json:"accounts"`
	InvitedAccounts []TeamAcc `json:"invited_accounts"`
}

// TeamAcc a team member.
type TeamAcc struct {
	AccountID    string `json:"account_id"`
	EmailAddress string `json:"email_address"`
	RoleCode     string `json:"role_code"`
}

type teamRaw struct {
	Team Team `json:"team"`
}

// Get returns information about your Team as well as a list of its members. If you do not belong to a Team,
// a 404 error with an error_name of "not_found" will be returned.
func (c *TeamAPI) Get() (*Team, error) {
	team := &teamRaw{}
	if err := c.getAndParse("team", nil, team); err != nil {
		return nil, err
	}
	return &team.Team, nil
}

// Create creates a new Team and makes you a member. You must not currently belong to a Team to invoke.
func (c *TeamAPI) Create(name string) (*Team, error) {
	team := &teamRaw{}
	err := c.postFormAndParse("team/create", &struct {
		Name string `form:"name"`
	}{
		Name: name,
	}, team)
	return &team.Team, err
}

// Update updates the name of your Team.
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

// Delete deletes your Team. Can only be invoked when you have a Team with only one member (yourself).
func (c *TeamAPI) Delete() (ok bool, err error) {
	resp, err := c.postForm("team/destroy", nil)
	if err != nil {
		return false, err
	}
	defer func() { err = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	return true, nil
}

type teamPostArgs struct {
	AccountID    *string `form:"account_id,omitempty"`
	EmailAddress *string `form:"email_address,omitempty"`
}

func (c *TeamAPI) addOrRemoveUser(ept string, accountID, emailAddress *string) (*Team, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	team := &teamRaw{}
	err := c.postFormAndParse(ept, &teamPostArgs{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, team)
	return &team.Team, err
}

// AddUser adds or invites a user (specified using the EmailAddress parameter) to your Team. If the user does not
// currently have a HelloSign Account, a new one will be created for them. If the user currently has a paid subscription,
// they will not automatically join the Team but instead will be sent an invitation to join. If a user is already a
// part of another Team, a "team_invite_failed" error will be returned.
func (c *TeamAPI) AddUser(accountID, emailAddress *string) (*Team, error) {
	return c.addOrRemoveUser("team/add_member", accountID, emailAddress)
}

// RemoveUser removes a user from your Team. If the user had an outstanding invitation to your Team the invitation will be expired.
func (c *TeamAPI) RemoveUser(accountID, emailAddress *string) (*Team, error) {
	return c.addOrRemoveUser("team/remove_member", accountID, emailAddress)
}
