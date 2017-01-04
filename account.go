// Copyright 2016 Precisely AB.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package hellosign

// AccountAPI used for account manipulations.
type AccountAPI struct {
	*hellosign
}

// NewAccountAPI creates a new api client for account operations.
func NewAccountAPI(apiKey string) *AccountAPI {
	return &AccountAPI{newHellosign(apiKey)}
}

// Acc contains information about an account and its settings.
type Acc struct {
	AccountID    string  `json:"account_id"`
	EmailAddress string  `json:"email_address"`
	CallbackURL  *string `json:"callback_url"`
	IsPaidHS     bool    `json:"is_paid_hs"`
	IsPaidHF     bool    `json:"is_paid_hf"`
	Quotas       struct {
		TemplatesLeft            *uint64 `json:"templates_left"`
		APISignatureRequestsLeft *uint64 `json:"api_signature_requests_left"`
		DocumentsLeft            *uint64 `json:"documents_left"`
	} `json:"quotas"`
	RoleCode *string `json:"role_code"`
}

type accRaw struct {
	Account Acc `json:"account"`
}

// Get returns your Account settings.
func (c *AccountAPI) Get() (*Acc, error) {
	acc := &accRaw{}
	if err := c.getAndParse("account", nil, acc); err != nil {
		return nil, err
	}
	return &acc.Account, nil
}

// Update sets your account settings.
func (c *AccountAPI) Update(callbackURL string) (*Acc, error) {
	acc := &accRaw{}
	err := c.postFormAndParse("account", &struct {
		CallbackURL string `form:"callback_url"`
	}{
		CallbackURL: callbackURL,
	}, acc)
	return &acc.Account, err
}

func (c *AccountAPI) createOrVerify(ept, emailAddress string) (*Acc, error) {
	acc := &accRaw{}
	err := c.postFormAndParse(ept, &struct {
		EmailAddress string `form:"email_address"`
	}{
		EmailAddress: emailAddress,
	}, acc)
	return &acc.Account, err
}

// Create signs up for a new HelloSign Account.
func (c *AccountAPI) Create(emailAddress string) (*Acc, error) {
	return c.createOrVerify("account/create", emailAddress)
}

// Verify whether a HelloSign Account exists.
func (c *AccountAPI) Verify(emailAddress string) (*Acc, error) {
	return c.createOrVerify("account/verify", emailAddress)
}
