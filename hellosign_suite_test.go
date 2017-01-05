package hellosign_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
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

func parseRequestParameters(req *http.Request) (map[string]string, error) {
	var params map[string]string
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return params, err
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return params, fmt.Errorf("invalid media type")
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return params, err
	}
	mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return params, err
		}
		key := p.FormName()
		b, err := ioutil.ReadAll(p)
		if err != nil {
			return params, err
		}
		params[key] = string(b)
	}
	return params, nil
}
