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

type SignatureRequestAPI struct {
	*hellosign
}

func NewSignatureRequestAPI(apiKey string) *SignatureRequestAPI {
	return &SignatureRequestAPI{newHellosign(apiKey)}
}

type SigReq struct {
	SignatureRequestID string        `json:"signature_request_id"`
	Title              string        `json:"title"`
	Subject            string        `json:"subject"`
	Message            string        `json:"message"`
	IsComplete         bool          `json:"is_complete"`
	IsDeclined         bool          `json:"is_declined"`
	HasError           bool          `json:"has_error"`
	CustomFields       []interface{} `json:"custom_fields"`
	ResponseData       []struct {
		ApiID       string `json:"api_id"`
		Name        string `json:"name"`
		SignatureId string `json:"signature_id"`
		Value       bool   `json:"value"`
		Type        string `json:"type"`
	} `json:"response_data"`
	SigningURL            *string `json:"signing_url"`
	SigningRedirectURL    *string `json:"signing_redirect_url"`
	DetailsURL            string  `json:"details_url"`
	RequesterEmailAddress string  `json:"requester_email_address"`
	Signatures            []struct {
		SignatureID        string  `json:"signature_id"`
		SignerEmailAddress string  `json:"signer_email_address"`
		SignerName         string  `json:"signer_name"`
		Order              *uint64 `json:"order"`
		StatusCode         string  `json:"status_code"`
		SignedAt           *uint64 `json:"signed_at"`
		LastViewedAt       *uint64 `json:"last_viewed_at"`
		LastRemindedAt     *uint64 `json:"last_reminded_at"`
		HasPin             bool    `json:"has_pin"`
	} `json:"signatures"`
	CCEmailAddresses []string `json:"cc_email_addresses"`
}

type sigReqRaw struct {
	SigReq SigReq `json:"signature_request"`
}

func (c *SignatureRequestAPI) Get(signatureRequestID string) (*SigReq, error) {
	resp, err := c.get(fmt.Sprintf("signature_request/%s", signatureRequestID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

type SigReqLst struct {
	ListInfo struct {
		Page       uint64 `json:"page"`
		NumPages   uint64 `json:"num_pages"`
		NumResults uint64 `json:"num_results"`
		PageSize   uint64 `json:"page_size"`
	} `json:"list_info"`
	SignatureRequests []SigReq `json:"signature_requests"`
}

type SigReqLstParms struct {
	AccountId *string `form:"account_id,omitempty"`
	Page      *uint64 `form:"page,omitempty"`
	PageSize  *uint64 `form:"page_size,omitempty"`
	Query     *string `form:"query,omitempty"`
}

func (c *SignatureRequestAPI) List(parms SigReqLstParms) (*SigReqLst, error) {
	paramString, err := form.EncodeToString(parms)
	if err != nil {
		return nil, err
	}
	resp, err := c.get("signature_request/list", &paramString)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	lst := &SigReqLst{}
	if err := c.parseResponse(resp, lst); err != nil {
		return nil, err
	}
	return lst, nil
}

type SigReqSendParms struct {
	TestMode              int8                    `form:"test_mode,omitempty"`
	AllowDecline          int8                    `form:"allow_decline,omitempty"`
	File                  [][]byte                `form:"file,omitempty"`
	FileUrl               []string                `form:"file_url,omitempty"`
	Title                 string                  `form:"title,omitempty"`
	Subject               string                  `form:"subject,omitempty"`
	Message               string                  `form:"message,omitempty"`
	SigningRedirectUrl    string                  `form:"message,omitempty"`
	Signers               []SigReqSendParmsSigner `form:"signers"`
	CCEmailAddresses      []string                `form:"cc_email_addresses,omitempty"`
	UseTextTags           int8                    `form:"use_text_tags,omitempty"`
	HideTextTags          int8                    `form:"hide_text_tags,omitempty"`
	Metadata              map[string]string       `form:"metadata,omitempty"`
	ClientId              string                  `form:"client_id,omitempty"`
	FormFieldsPerDocument string                  `form:"form_fields_per_documents,omitempty"`
}

type SigReqSendParmsSigner struct {
	Name         string  `form:"name"`
	EmailAddress string  `form:"email_address"`
	Order        *uint64 `form:"order,omitempty"`
	Pin          string  `form:"pin,omitempty"`
}

func (c *SignatureRequestAPI) Send(parms SigReqSendParms) (*SigReq, error) {
	if len(parms.File) == 0 && len(parms.FileUrl) == 0 {
		return nil, errors.New("Specify either file or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileUrl) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	resp, err := c.postForm("signature_request/send", &parms)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

type SigReqSendTplParms struct {
	TestMode           int8                                `form:"test_mode,omitempty"`
	AllowDecline       int8                                `form:"allow_decline,omitempty"`
	TemplateId         string                              `form:"template_id,omitempty"`
	TemplateIds        []string                            `form:"template_ids,omitempty"`
	Title              string                              `form:"title,omitempty"`
	Subject            string                              `form:"subject,omitempty"`
	Message            string                              `form:"message,omitempty"`
	SigningRedirectUrl string                              `form:"signing_redirect_url,omitempty"`
	Signers            map[string]SigReqSendTplParmsSigner `form:"signers"`
	Ccs                map[string]SigReqSendTplParmsCcs    `form:"ccs,omitempty"`
	CustomFields       string                              `form:"custom_fields,omitempty"`
	Metadata           map[string]string                   `form:"metadata,omitempty"`
	ClientId           string                              `form:"client_id,omitempty"`
}

type SigReqSendTplParmsSigner struct {
	Name         string `form:"name"`
	EmailAddress string `form:"email_address"`
	Pin          string `form:"pin,omitempty"`
}

type SigReqSendTplParmsCcs struct {
	EmailAddress string `form:"email_address"`
}

func (c *SignatureRequestAPI) SendWithTemplate(parms SigReqSendTplParms) (*SigReq, error) {
	if parms.TemplateId == "" && len(parms.TemplateIds) == 0 {
		return nil, errors.New("Specify either template id or template ids, none given")
	}
	if parms.TemplateId != "" && len(parms.TemplateIds) > 0 {
		return nil, errors.New("Specify either template id or template ids, both given")
	}
	resp, err := c.postForm("signature_request/send_with_template", &parms)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

func (c *SignatureRequestAPI) SendReminder(signatureRequestID, emailAddress string, name *string) (*SigReq, error) {
	resp, err := c.postForm(fmt.Sprintf("signature_request/remind/%s", signatureRequestID), &struct {
		EmailAddress string  `form:"email_address"`
		Name         *string `form:"name,omitempty"`
	}{
		EmailAddress: emailAddress,
		Name:         name,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

func (c *SignatureRequestAPI) Update(signatureRequestID, signatureId, email string) (*SigReq, error) {
	resp, err := c.postForm(fmt.Sprintf("signature_request/update/%s", signatureRequestID), &struct {
		SignatureId  string `form:"signature_id"`
		EmailAddress string `form:"email_address"`
	}{
		SignatureId:  signatureId,
		EmailAddress: email,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

func (c *SignatureRequestAPI) Cancel(signatureRequestID string) (bool, error) {
	resp, err := c.postForm(fmt.Sprintf("signature_request/cancel/%s", signatureRequestID), nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	return true, nil
}

type FileURL struct {
	FileUrl   string `json:"file_url"`
	ExpiresAt uint64 `json:"expires_at"`
}

func (c *SignatureRequestAPI) Files(signatureRequestID, fileType string, getURL bool) (*[]byte, *FileURL, error) {
	return c.getFiles(fmt.Sprintf("signature_request/files/%", signatureRequestID), fileType, getURL)
}

type SigReqEmbSendParms struct {
	TestMode              int8                       `form:"test_mode,omitempty"`
	AllowDecline          int8                       `form:"allow_decline,omitempty"`
	ClientId              string                     `form:"client_id"`
	File                  [][]byte                   `form:"file,omitempty"`
	FileUrl               []string                   `form:"file_url,omitempty"`
	Title                 string                     `form:"title,omitempty"`
	Subject               string                     `form:"subject,omitempty"`
	Message               string                     `form:"message,omitempty"`
	Signers               []SigReqEmbSendParmsSigner `form:"signers"`
	CCEmailAddresses      []string                   `form:"cc_email_addresses,omitempty"`
	UseTextTags           int8                       `form:"use_text_tags,omitempty"`
	HideTextTags          int8                       `form:"hide_text_tags,omitempty"`
	Metadata              map[string]string          `form:"metadata,omitempty"`
	FormFieldsPerDocument string                     `form:"form_fields_per_documents,omitempty"`
}

type SigReqEmbSendParmsSigner struct {
	Name         string  `form:"name"`
	EmailAddress string  `form:"email_address"`
	Order        *uint64 `form:"order,omitempty"`
	Pin          string  `form:"pin,omitempty"`
}

func (c *SignatureRequestAPI) SendEmbedded(parms SigReqEmbSendParms) (*SigReq, error) {
	if len(parms.File) == 0 && len(parms.FileUrl) == 0 {
		return nil, errors.New("Specify either file or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileUrl) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	resp, err := c.postForm("signature_request/create_embedded", &parms)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sigReq := &sigReqRaw{}
	if err := c.parseResponse(resp, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}
