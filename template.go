// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"fmt"
	"net/http"
)

// TemplateAPI used for manipulating templates.
type TemplateAPI struct {
	*hellosign
}

// NewTemplateAPI creates a new api client for template endpoints.
func NewTemplateAPI(apiKey string) *TemplateAPI {
	return &TemplateAPI{newHellosign(apiKey)}
}

// Tpl contains information about the templates you and your team have created
type Tpl struct {
	TemplateID  string `json:"template_id"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	SignerRoles []struct {
		Name  string  `json:"name"`
		Order *uint64 `json:"order"`
	} `json:"signer_roles"`
	CCRoles []struct {
		Name string `json:"name"`
	}
	Documents []struct {
		Index        uint64      `json:"index"`
		Name         string      `json:"name"`
		FormFields   []FormField `json:"form_fields"`
		CustomFields []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"custom_fields"`
	} `json:"documents"`
	Accounts []struct {
		AccountID    string `json:"account_id"`
		EmailAddress string `json:"email_address"`
	} `json:"accounts"`
}

type tplRaw struct {
	Template Tpl `json:"template"`
}

// Get returns the Template specified by the id parameter.
func (c *TemplateAPI) Get(templateID string) (*Tpl, error) {
	tpl := &tplRaw{}
	if err := c.getAndParse(fmt.Sprintf("template/%s", templateID), nil, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}

// TplLst is a list of templates accessible by this account.
type TplLst struct {
	ListInfo  ListInfo `json:"list_info"`
	Templates []Tpl    `json:"templates"`
}

// List returns a list of the Templates that are accessible by you.
func (c *TemplateAPI) List(parms ListParms) (*TplLst, error) {
	lst := &TplLst{}
	err := c.list("template/list", parms, lst)
	return lst, err
}

type tplAddRemParms struct {
	AccountID    *string `form:"account_id,omitempty"`
	EmailAddress *string `form:"email_address,omitempty"`
}

func (c *TemplateAPI) addRemove(ept string, accountID, emailAddress *string) (*Tpl, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	tpl := &tplRaw{}
	if err := c.postFormAndParse(ept, &tplAddRemParms{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}

// AddUser gives the specified Account access to the specified Template. The specified Account must be a part of your Team.
func (c *TemplateAPI) AddUser(templateID string, accountID, emailAddress *string) (*Tpl, error) {
	return c.addRemove(fmt.Sprintf("template/add_user/%s", templateID), accountID, emailAddress)
}

// RemoveUser removes the specified Account's access to the specified Template.
func (c *TemplateAPI) RemoveUser(templateID string, accountID, emailAddress *string) (*Tpl, error) {
	return c.addRemove(fmt.Sprintf("template/remove_user/%s", templateID), accountID, emailAddress)
}

// Files obtain a copy of the original files specified by the template_id parameter.
func (c *TemplateAPI) Files(templateID, fileType string, getURL bool) ([]byte, *FileURL, error) {
	return c.getFiles(fmt.Sprintf("template/files/%s", templateID), fileType, getURL)
}

// Delete completely deletes the template specified from the account.
func (c *TemplateAPI) Delete(templateID string) (ok bool, err error) {
	return c.postEmptyExpect(fmt.Sprintf("template/delete/%s", templateID), http.StatusOK)
}

// TplEmbCreateParms parameters for creating template drafts.
type TplEmbCreateParms struct {
	TestMode    int8               `form:"test_mode,omitempty"`
	ClientID    string             `form:"client_id"`
	File        [][]byte           `form:"file,omitempty"`
	FileURL     []string           `form:"file_url,omitempty"`
	Title       string             `form:"title,omitempty"`
	Subject     string             `form:"subject,omitempty"`
	Message     string             `form:"message,omitempty"`
	SignerRoles []TplEmbSignerRole `form:"signer_roles,omitempty"`
	CCRoles     []string           `form:"cc_email_addresses,omitempty"`
	MergeFields []TplEmbMergeField `form:"merge_fields,omitempty"`
	Metadata    map[string]string  `form:"metadata,omitempty"`
}

// TplEmbSignerRole role parameter for template.
type TplEmbSignerRole struct {
	Name  string  `form:"name"`
	Order *uint64 `form:"order,omitempty"`
}

// TplEmbMergeField the merge fields that can be placed on the template's document(s) by the user claiming the template draft.
type TplEmbMergeField struct {
	Name string `form:"name"`
	Type string `form:"type"`
}

// CreateEmbeddedDraft he first step in an embedded template workflow. Creates a draft template
// that can then be further set up in the template 'edit' stage.
func (c *TemplateAPI) CreateEmbeddedDraft(parms TplEmbCreateParms) (*Tpl, error) {
	if len(parms.File) == 0 && len(parms.FileURL) == 0 {
		return nil, errors.New("Specify either file or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileURL) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	tpl := &tplRaw{}
	if err := c.postFormAndParse("template/create_embedded_draft", parms, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}
