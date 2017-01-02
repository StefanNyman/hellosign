package hellosign_test

import (
	"errors"
	"hellosign"
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/jarcoal/httpmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	callbackURL string = "http://thisisurl.com/hellosign/account_callback"
)

var _ = Describe("Account", func() {
	var (
		client            *hellosign.AccountAPI
		apiKey            string
		getAccountResp    string
		getAccountRespRaw string
	)

	_ = BeforeEach(func() {
		apiKey = "asdf"
		client = hellosign.NewAccountAPI(apiKey)
		getAccountResp = `
			{
        "account_id": "5008b25c7f67153e57d5a357b1687968068fb465",
        "email_address": "me@hellosign.com",
        "is_paid_hs" : true,
        "is_paid_hf" : false,
        "quotas" : {
            "api_signature_requests_left": 1250,
            "documents_left": null,
            "templates_left": null
        },
        "callback_url": null,
        "role_code": null
			}
		`
		getAccountRespRaw = `
			{
		    "account": ` + getAccountResp + `
			}
		`
	})

	It("fetches an account", func() {
		httpmock.RegisterResponder(http.MethodGet, hellosign.GetEptURL("account"),
			httpmock.NewStringResponder(http.StatusOK, getAccountRespRaw))
		acc, err := client.Get()
		Expect(err).To(BeNil())
		Expect(acc).ToNot(BeNil())
		Expect(acc.AccountID).To(Equal("5008b25c7f67153e57d5a357b1687968068fb465"))
		Expect(acc.EmailAddress).To(Equal("me@hellosign.com"))
		Expect(acc.IsPaidHS).To(BeTrue())
		Expect(acc.IsPaidHF).To(BeFalse())
		Expect(*acc.Quotas.APISignatureRequestsLeft).To(Equal(uint64(1250)))
		Expect(acc.Quotas.DocumentsLeft).To(BeNil())
		Expect(acc.Quotas.TemplatesLeft).To(BeNil())
		Expect(acc.CallbackURL).To(BeNil())
		Expect(acc.RoleCode).To(BeNil())
	})

	It("creates an account", func() {
		httpmock.RegisterResponder(http.MethodPost, hellosign.GetEptURL("account/create"),
			mockResponseCreateVerify)
		addr := "test@test.com"
		acc, err := client.Create(addr)
		Expect(err).To(BeNil())
		Expect(acc.EmailAddress).To(Equal(addr))
	})

	It("updates an account", func() {
		httpmock.RegisterResponder(http.MethodPost, hellosign.GetEptURL("account"),
			mockResponseUpdate)
		acc, err := client.Update(callbackURL)
		Expect(err).To(BeNil())
		Expect(*acc.CallbackURL).To(Equal(callbackURL))
	})

	It("verifies an account", func() {
		httpmock.RegisterResponder(http.MethodPost, hellosign.GetEptURL("account/verify"),
			mockResponseCreateVerify)
		addr := "test@test.com"
		acc, err := client.Verify(addr)
		Expect(err).To(BeNil())
		Expect(acc.EmailAddress).To(Equal(addr))
	})

})

func mockResponseCreateVerify(req *http.Request) (*http.Response, error) {
	b, _ := ioutil.ReadAll(req.Body)
	vals, err := url.ParseQuery(string(b))
	if err != nil {
		return nil, err
	}
	emailAddress, found := vals["email_address"]
	if !found {
		return nil, errors.New("email_address not found in values")
	}
	return httpmock.NewJsonResponse(http.StatusOK, &struct {
		Account struct {
			EmailAddress string `json:"email_address"`
		} `json:"account"`
	}{
		Account: struct {
			EmailAddress string `json:"email_address"`
		}{
			EmailAddress: emailAddress[0],
		},
	})
}

func mockResponseUpdate(req *http.Request) (*http.Response, error) {
	b, _ := ioutil.ReadAll(req.Body)
	vals, err := url.ParseQuery(string(b))
	if err != nil {
		return nil, err
	}
	cbURLArr, found := vals["callback_url"]
	if !found {
		return nil, errors.New("callback_url not found in values")
	}
	Expect(cbURLArr[0]).To(Equal(callbackURL))
	// this is not very nice... need better abstraction
	return httpmock.NewJsonResponse(http.StatusOK, &struct {
		Account struct {
			CallbackURL string `json:"callback_url"`
		} `json:"account"`
	}{
		Account: struct {
			CallbackURL string `json:"callback_url"`
		}{
			CallbackURL: cbURLArr[0],
		},
	})
}
