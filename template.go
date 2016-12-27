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

type TemplateAPI struct {
	*hellosign
}

func NewTemplateAPI(apiKey string) *TemplateAPI {
	return &TemplateAPI{newHellosign(apiKey)}
}

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
		Index      uint64 `json:"index"`
		Name       string `json:"name"`
		FormFields []struct {
			ApiID    string `json:"api_id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			X        uint64 `json:"x"`
			Y        uint64 `json:"y"`
			Width    uint64 `json:"width"`
			Height   uint64 `json:"height"`
			Required bool   `json:"required"`
		} `json:"form_fields"`
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

func (c *TemplateAPI) Get(templateID string) (*Tpl, error) {
	tpl := &tplRaw{}
	if err := c.getAndParse(fmt.Sprintf("template/%s", templateID), nil, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}

type TplLst struct {
	ListInfo struct {
		Page       uint64 `json:"page"`
		NumPages   uint64 `json:"num_pages"`
		NumResults uint64 `json:"num_results"`
		PageSize   uint64 `json:"page_size"`
	} `json:"list_info"`
	Templates []Tpl `json:"templates"`
}

type TplLstParms struct {
	AccountId *string `form:"account_id,omitempty"`
	Page      *uint64 `form:"page,omitempty"`
	PageSize  *uint64 `form:"page_size,omitempty"`
	Query     *string `form:"query,omitempty"`
}

func (c *TemplateAPI) List(parms TplLstParms) (*TplLst, error) {
	paramString, err := form.EncodeToString(parms)
	if err != nil {
		return nil, err
	}
	lst := &TplLst{}
	if err := c.getAndParse("template/list", &paramString, lst); err != nil {
		return nil, err
	}
	return lst, nil
}

func (c *TemplateAPI) AddUser(templateID string, accountID, emailAddress *string) (*Tpl, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	tpl := &tplRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("template/add_user/%s", templateID), &struct {
		AccountID    *string `form:"account_id,omitempty"`
		EmailAddress *string `form:"email_address,omitempty"`
	}{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}

func (c *TemplateAPI) RemoveUser(templateID string, accountID, emailAddress *string) (*Tpl, error) {
	if accountID != nil && emailAddress != nil {
		return nil, errors.New("Specify either account id or email address, both given")
	}
	tpl := &tplRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("template/remove_user/%s", templateID), &struct {
		AccountID    *string `form:"account_id,omitempty"`
		EmailAddress *string `form:"email_address,omitempty"`
	}{
		AccountID:    accountID,
		EmailAddress: emailAddress,
	}, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}

func (c *TemplateAPI) Files(templateID, fileType string, getURL bool) ([]byte, *FileURL, error) {
	return c.getFiles(fmt.Sprintf("template/files/%s", templateID), fileType, getURL)
}

func (c *TemplateAPI) Delete(templateID string) (bool, error) {
	resp, err := c.postForm(fmt.Sprintf("template/delete/%s", templateID), nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	return true, nil
}

type TplEmbCreateParms struct {
	TestMode    int8               `form:"test_mode,omitempty"`
	ClientId    string             `form:"client_id"`
	File        [][]byte           `form:"file,omitempty"`
	FileUrl     []string           `form:"file_url,omitempty"`
	Title       string             `form:"title,omitempty"`
	Subject     string             `form:"subject,omitempty"`
	Message     string             `form:"message,omitempty"`
	SignerRoles []TplEmbSignerRole `form:"signer_roles,omitempty"`
	CCRoles     []string           `form:"cc_email_addresses,omitempty"`
	MergeFields []TplEmbMergeField `form:"merge_fields,omitempty"`
	Metadata    map[string]string  `form:"metadata,omitempty"`
}

type TplEmbSignerRole struct {
	Name  string  `form:"name"`
	Order *uint64 `form:"order,omitempty"`
}

type TplEmbMergeField struct {
	Name string `form:"name"`
	Type string `form:"type"`
}

func (c *TemplateAPI) CreateEmbeddedDraft(parms TplEmbCreateParms) (*Tpl, error) {
	if len(parms.File) == 0 && len(parms.FileUrl) == 0 {
		return nil, errors.New("Specify either file or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileUrl) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	tpl := &tplRaw{}
	if err := c.postFormAndParse("template/create_embedded_draft", parms, tpl); err != nil {
		return nil, err
	}
	return &tpl.Template, nil
}
