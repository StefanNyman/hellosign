package hellosign_test

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"unicode"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/jarcoal/httpmock.v1"

	"testing"
)

func DumpRequest(req *http.Request) {
	d, err := httputil.DumpRequest(req, true)
	if err == nil {
		fmt.Println(string(d))
	}
}

func ReplaceWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func StringsNonWhitespaceEqual(a, b string) bool {
	return ReplaceWhitespace(a) == ReplaceWhitespace(b)
}

var _ = BeforeSuite(func() {
	// block all HTTP requests
	httpmock.Activate()
})

var _ = BeforeEach(func() {
	// remove any mocks
	httpmock.Reset()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestHellosign(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hellosign Suite")
}
