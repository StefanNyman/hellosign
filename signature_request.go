// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// SignatureRequestAPI used for signature request manipulations.
type SignatureRequestAPI struct {
	*hellosign
}

// NewSignatureRequestAPI creates a new api client for signature request manipulations.
func NewSignatureRequestAPI(apiKey string) *SignatureRequestAPI {
	return &SignatureRequestAPI{newHellosign(apiKey)}
}

// SigReq contains information regarding documents that need to be signed.
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
		APIID       string `json:"api_id"`
		Name        string `json:"name"`
		SignatureID string `json:"signature_id"`
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

// Get returns the status of the SignatureRequest specified by the signatureRequestID parameter.
func (c *SignatureRequestAPI) Get(signatureRequestID string) (*SigReq, error) {
	sigReq := &sigReqRaw{}
	err := c.getAndParse(fmt.Sprintf("signature_request/%s", signatureRequestID), nil, sigReq)
	return &sigReq.SigReq, err
}

// SigReqLst is a list of signature requests managed by this account.
type SigReqLst struct {
	ListInfo          ListInfo `json:"list_info"`
	SignatureRequests []SigReq `json:"signature_requests"`
}

// List returns a list of SignatureRequests that you can access. This includes SignatureRequests
// you have sent as well as received, but not ones that you have been CCed on.
func (c *SignatureRequestAPI) List(parms ListParms) (*SigReqLst, error) {
	lst := &SigReqLst{}
	err := c.list("signature_request/list", parms, lst)
	return lst, err
}

// SigReqSendParms parameters for creating a new signature request.
type SigReqSendParms struct {
	File                  [][]byte                `form:"file,omitempty"`
	FileURL               []string                `form:"file_url,omitempty"`
	FileIO                []io.Reader             `form:"-"`
	Title                 string                  `form:"title,omitempty"`
	Subject               string                  `form:"subject,omitempty"`
	Message               string                  `form:"message,omitempty"`
	SigningRedirectURL    string                  `form:"message,omitempty"`
	Signers               []SigReqSendParmsSigner `form:"signers"`
	CCEmailAddresses      []string                `form:"cc_email_addresses,omitempty"`
	Metadata              map[string]string       `form:"metadata,omitempty"`
	ClientID              string                  `form:"client_id,omitempty"`
	FormFieldsPerDocument string                  `form:"form_fields_per_documents,omitempty"`
	TestMode              int8                    `form:"test_mode,omitempty"`
	AllowDecline          int8                    `form:"allow_decline,omitempty"`
	UseTextTags           int8                    `form:"use_text_tags,omitempty"`
	HideTextTags          int8                    `form:"hide_text_tags,omitempty"`
}

// SigReqSendParmsSigner represents a person that should sign a document. Each signer must be unique.
type SigReqSendParmsSigner struct {
	Name         string  `form:"name"`
	EmailAddress string  `form:"email_address"`
	Order        *uint64 `form:"order,omitempty"`
	Pin          string  `form:"pin,omitempty"`
}

// Send creates and sends a new SignatureRequest with the submitted documents. If FormFieldsPerDocument is
// not specified, a signature page will be affixed where all signers will be required to add their signature,
// signifying their agreement to all contained documents.
func (c *SignatureRequestAPI) Send(parms SigReqSendParms) (*SigReq, error) {
	if len(parms.File) == 0 && len(parms.FileURL) == 0 {
		return nil, errors.New("Specify either file or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileURL) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	if len(parms.File) > 0 && len(parms.FileIO) > 0 {
		return nil, errors.New("Specify either file or file io, both given")
	}
	if len(parms.FileIO) > 0 && len(parms.FileURL) > 0 {
		return nil, errors.New("Specify either file io or file url, both given")
	}
	if len(parms.FileIO) > 0 {
		for _, f := range parms.FileIO {
			fc, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			parms.File = append(parms.File, fc)
		}
	}
	sigReq := &sigReqRaw{}
	if err := c.postFormAndParse("signature_request/send", &parms, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

// SigReqSendTplParms parameters for creating a signature request from a template.
type SigReqSendTplParms struct {
	TemplateID         string                              `form:"template_id,omitempty"`
	TemplateIds        []string                            `form:"template_ids,omitempty"`
	Title              string                              `form:"title,omitempty"`
	Subject            string                              `form:"subject,omitempty"`
	Message            string                              `form:"message,omitempty"`
	SigningRedirectURL string                              `form:"signing_redirect_url,omitempty"`
	Signers            map[string]SigReqSendTplParmsSigner `form:"signers"`
	Ccs                map[string]SigReqSendTplParmsCcs    `form:"ccs,omitempty"`
	CustomFields       string                              `form:"custom_fields,omitempty"`
	Metadata           map[string]string                   `form:"metadata,omitempty"`
	ClientID           string                              `form:"client_id,omitempty"`
	TestMode           int8                                `form:"test_mode,omitempty"`
	AllowDecline       int8                                `form:"allow_decline,omitempty"`
}

// SigReqSendTplParmsSigner represents a person that should sign the template signature request.
type SigReqSendTplParmsSigner struct {
	Name         string `form:"name"`
	EmailAddress string `form:"email_address"`
	Pin          string `form:"pin,omitempty"`
}

// SigReqSendTplParmsCcs an email address that should be cc'd when the template signature
// request is signed.
type SigReqSendTplParmsCcs struct {
	EmailAddress string `form:"email_address"`
}

// SendWithTemplate creates and sends a new SignatureRequest based off of the Template specified with the TemplateID parameter.
func (c *SignatureRequestAPI) SendWithTemplate(parms SigReqSendTplParms) (*SigReq, error) {
	if parms.TemplateID == "" && len(parms.TemplateIds) == 0 {
		return nil, errors.New("Specify either template id or template ids, none given")
	}
	if parms.TemplateID != "" && len(parms.TemplateIds) > 0 {
		return nil, errors.New("Specify either template id or template ids, both given")
	}
	sigReq := &sigReqRaw{}
	if err := c.postFormAndParse("signature_request/send_with_template", &parms, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

// SendReminder sends an email to the signer reminding them to sign the signature request. You cannot send a
// reminder within 1 hour of the last reminder that was sent. This includes manual AND automatic reminders.
func (c *SignatureRequestAPI) SendReminder(signatureRequestID, emailAddress string, name *string) (*SigReq, error) {
	sigReq := &sigReqRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("signature_request/remind/%s", signatureRequestID), &struct {
		EmailAddress string  `form:"email_address"`
		Name         *string `form:"name,omitempty"`
	}{
		EmailAddress: emailAddress,
		Name:         name,
	}, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

// Update updates the email address for a given signer on a signature request. You can listen for the
// "signature_request_email_bounce" event on your app or account to detect bounced emails, and respond with this method.
func (c *SignatureRequestAPI) Update(signatureRequestID, signatureID, email string) (*SigReq, error) {
	sigReq := &sigReqRaw{}
	if err := c.postFormAndParse(fmt.Sprintf("signature_request/update/%s", signatureRequestID), &struct {
		SignatureID  string `form:"signature_id"`
		EmailAddress string `form:"email_address"`
	}{
		SignatureID:  signatureID,
		EmailAddress: email,
	}, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}

// Cancel Queues a SignatureRequest to be canceled. The cancelation is asynchronous and a successful call to this endpoint
// will return a 200 OK response if the signature request is eligible to be canceled and has been successfully queued. To be
// eligible for cancelation, a signature request must have been sent successfully and must be unsigned. Once canceled, signers
// will not be able to sign the signature request or access its documents. Canceling a signature request is not reversible.
//
// Configuring a callback handler to listen for the SIGNATURE_REQUEST_CANCELED event is recommended to receive a notification
// when the cancelation has taken place. If a callback handler has been configured and this event has not been received within
// 60 minutes of making the call, please check the status of the request in the API Dashboard and retry the request if necessary.
func (c *SignatureRequestAPI) Cancel(signatureRequestID string) (ok bool, err error) {
	return c.postEmptyExpect(fmt.Sprintf("signature_request/cancel/%s", signatureRequestID), http.StatusOK)
}

// FileURL is an URL with an expiration time.
type FileURL struct {
	FileURL   string `json:"file_url"`
	ExpiresAt uint64 `json:"expires_at"`
}

// Files obtain a copy of the current documents specified by the signatureRequestID parameter.
func (c *SignatureRequestAPI) Files(signatureRequestID, fileType string, getURL bool) ([]byte, *FileURL, error) {
	return c.getFiles(fmt.Sprintf("signature_request/files/%s", signatureRequestID), fileType, getURL)
}

// SigReqEmbSendParms parameters for creating an embedded signature request.
type SigReqEmbSendParms struct {
	ClientID              string                     `form:"client_id"`
	File                  [][]byte                   `form:"file,omitempty"`
	FileIO                []io.Reader                `form:"-"`
	FileURL               []string                   `form:"file_url,omitempty"`
	Title                 string                     `form:"title,omitempty"`
	Subject               string                     `form:"subject,omitempty"`
	Message               string                     `form:"message,omitempty"`
	Signers               []SigReqEmbSendParmsSigner `form:"signers"`
	CCEmailAddresses      []string                   `form:"cc_email_addresses,omitempty"`
	Metadata              map[string]string          `form:"metadata,omitempty"`
	FormFieldsPerDocument string                     `form:"form_fields_per_documents,omitempty"`
	UseTextTags           int8                       `form:"use_text_tags,omitempty"`
	HideTextTags          int8                       `form:"hide_text_tags,omitempty"`
	TestMode              int8                       `form:"test_mode,omitempty"`
	AllowDecline          int8                       `form:"allow_decline,omitempty"`
}

// SigReqEmbSendParmsSigner represents a person that should sign a document. Each signer must be unique.
type SigReqEmbSendParmsSigner struct {
	Name         string  `form:"name"`
	EmailAddress string  `form:"email_address"`
	Order        *uint64 `form:"order,omitempty"`
	Pin          string  `form:"pin,omitempty"`
}

// SendEmbedded creates a new SignatureRequest with the submitted documents to be signed in an embedded iFrame.
// If FormFieldsPerDocument is not specified, a signature page will be affixed where all signers will be required to
// add their signature, signifying their agreement to all contained documents. Note that embedded signature requests
// can only be signed in embedded iFrames whereas normal signature requests can only be signed on HelloSign.
func (c *SignatureRequestAPI) SendEmbedded(parms SigReqEmbSendParms) (*SigReq, error) {
	if len(parms.File) == 0 && len(parms.FileURL) == 0 && len(parms.FileIO) == 0 {
		return nil, errors.New("Specify either file, file io or file url, none given")
	}
	if len(parms.File) > 0 && len(parms.FileURL) > 0 {
		return nil, errors.New("Specify either file or file url, both given")
	}
	if len(parms.File) > 0 && len(parms.FileIO) > 0 {
		return nil, errors.New("Specify either file or file io, both given")
	}
	if len(parms.FileIO) > 0 && len(parms.FileURL) > 0 {
		return nil, errors.New("Specify either file io or file url, both given")
	}
	if len(parms.FileIO) > 0 {
		for _, f := range parms.FileIO {
			fc, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			parms.File = append(parms.File, fc)
		}
	}
	sigReq := &sigReqRaw{}
	if err := c.postFormAndParse("signature_request/create_embedded", &parms, sigReq); err != nil {
		return nil, err
	}
	return &sigReq.SigReq, nil
}
