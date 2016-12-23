// Copyright 2016 Stefan Nyman.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

type AccountAPI struct {
	*hellosign
}

func NewAccountAPI(apiKey string) *AccountAPI {
	return &AccountAPI{newHellosign(apiKey)}
}

type Acc struct {
	AccountID    string  `json:"account_id"`
	EmailAddress string  `json:"email_address"`
	CallbackURL  *string `json:"callback_url"`
	IsPaidHS     bool    `json:"is_paid_hs"`
	IsPaidHF     bool    `json:"is_paid_hf"`
	Quotas       struct {
		TemplatesLeft            *uint64 `json:"templates_left"`
		ApiSignatureRequestsLeft *uint64 `json:"api_signature_requests_left"`
		DocumentsLeft            *uint64 `json:"documents_left"`
	} `json:"quotas"`
	RoleCode *string `json:"role_code"`
}

type accRaw struct {
	Account Acc `json:"account"`
}

func (c *AccountAPI) Get() (*Acc, error) {
	acc := &accRaw{}
	if err := c.getAndParse("account", nil, acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func (c *AccountAPI) Update(callbackURL string) (*Acc, error) {
	acc := &accRaw{}
	if err := c.postFormAndParse("account", &struct {
		CallbackURL string `form:"callback_url"`
	}{
		CallbackURL: callbackURL,
	}, acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func (c *AccountAPI) Create(emailAddress string) (*Acc, error) {
	acc := &accRaw{}
	if err := c.postFormAndParse("account/create", &struct {
		EmailAddress string `form:"email_address"`
	}{
		EmailAddress: emailAddress,
	}, acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

func (c *AccountAPI) Verify(emailAddress string) (*Acc, error) {
	acc := &accRaw{}
	if err := c.postFormAndParse("account/verify", &struct {
		EmailAddress string `form:"email_address"`
	}{
		EmailAddress: emailAddress,
	}, acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}
