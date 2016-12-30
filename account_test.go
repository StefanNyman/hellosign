package hellosign_test

import (
	"hellosign"
	"net/http"

	"gopkg.in/jarcoal/httpmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Expect(*acc.Quotas.ApiSignatureRequestsLeft).To(Equal(uint64(1250)))
		Expect(acc.Quotas.DocumentsLeft).To(BeNil())
		Expect(acc.Quotas.TemplatesLeft).To(BeNil())
		Expect(acc.CallbackURL).To(BeNil())
		Expect(acc.RoleCode).To(BeNil())
	})
})
