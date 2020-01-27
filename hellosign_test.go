package hellosign

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"unicode"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func dumpRequest(req *http.Request) {
	d, err := httputil.DumpRequest(req, true)
	if err == nil {
		fmt.Println(string(d))
	}
}

func replaceWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func stringsNonWhitespaceEqual(a, b string) bool {
	return replaceWhitespace(a) == replaceWhitespace(b)
}

var _ = Describe("Hellosign", func() {
	var (
		apiKey         string
		client         *hellosign
		getAccountResp string
		errorResponse  string
	)
	_ = BeforeEach(func() {
		apiKey = "asdf"
		getAccountResp = `
			{
		    "account": {
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
			}
		`
		errorResponse = `
			{
				"error": {
	        "error_msg": "Bad request",
	        "error_name": "bad_request"
	    	}
			}

		`
		client = newHellosign(apiKey)
	})

	It("sets correct client values", func() {
		Expect(client.apiKey).To(Equal(apiKey))
		Expect(client.LastStatusCode).To(Equal(0))
		Expect(client.RateLimit).To(Equal(uint64(0)))
		Expect(client.RateLimitRemaining).To(Equal(uint64(0)))
		Expect(client.RateLimitReset).To(Equal(uint64(0)))
	})

	It("generates correct urls", func() {
		accountURL := client.getEptURL("account")
		Expect(accountURL).To(Equal(fmt.Sprintf("%s/%s", baseURL, "account")))
	})

	It("performs a request", func() {
		httpmock.RegisterResponder(http.MethodGet, client.getEptURL("account"),
			httpmock.NewStringResponder(http.StatusOK, getAccountResp))
		req, _ := http.NewRequest(http.MethodGet, client.getEptURL("account"), nil)
		resp, err := client.perform(req)
		Expect(err).To(BeNil())
		uname, pwd, ok := req.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(pwd).To(Equal(""))
		Expect(uname).To(Equal(apiKey))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		b, _ := ioutil.ReadAll(resp.Body)
		Expect(stringsNonWhitespaceEqual(getAccountResp, string(b))).To(BeTrue())
	})

	It("parses response headers", func() {
		httpmock.RegisterResponder(http.MethodGet, client.getEptURL("account"),
			func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Header:     http.Header{},
				}
				resp.Header.Add(xRatelimitLimit, "3000")
				resp.Header.Add(xRatelimitLimitRemaining, "2999")
				resp.Header.Add(xRateLimitReset, "1")
				return resp, nil
			})
		_, err := client.get("account", nil)
		Expect(err).To(BeNil())
		Expect(client.RateLimit).To(Equal(uint64(3000)))
		Expect(client.RateLimitRemaining).To(Equal(uint64(2999)))
		Expect(client.RateLimitReset).To(Equal(uint64(1)))
	})

	It("produces errors on non 2xx responses", func() {
		httpmock.RegisterResponder(http.MethodGet, client.getEptURL("account"),
			httpmock.NewStringResponder(http.StatusBadRequest, errorResponse))
		req, _ := http.NewRequest(http.MethodGet, client.getEptURL("account"), nil)
		_, err := client.perform(req)
		Expect(err).ToNot(BeNil())
		hErr, ok := err.(APIErr)
		Expect(ok).To(BeTrue())
		Expect(hErr.Code).To(Equal(http.StatusBadRequest))
		Expect(hErr.Message).To(Equal("Bad request"))
		Expect(hErr.Name).To(Equal("bad_request"))
	})

	It("parses responses", func() {
		httpmock.RegisterResponder(http.MethodGet, client.getEptURL("account"),
			httpmock.NewStringResponder(http.StatusOK, getAccountResp))
		req, _ := http.NewRequest(http.MethodGet, client.getEptURL("account"), nil)
		resp, err := client.perform(req)
		Expect(err).To(BeNil())
		acc := &accRaw{}
		err = client.parseResponse(resp, acc)
		Expect(err).To(BeNil())
		account := acc.Account
		Expect(account.AccountID).To(Equal("5008b25c7f67153e57d5a357b1687968068fb465"))
		Expect(account.EmailAddress).To(Equal("me@hellosign.com"))
		Expect(account.IsPaidHS).To(BeTrue())
		Expect(account.IsPaidHF).To(BeFalse())
		Expect(*account.Quotas.APISignatureRequestsLeft).To(Equal(uint64(1250)))
		Expect(account.Quotas.DocumentsLeft).To(BeNil())
		Expect(account.Quotas.TemplatesLeft).To(BeNil())
		Expect(account.CallbackURL).To(BeNil())
		Expect(account.RoleCode).To(BeNil())
	})
})
