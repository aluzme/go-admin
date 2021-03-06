package echo

import (
	"net/http"
	"testing"

	"github.com/aluzme/go-admin/tests/common"
	"github.com/gavv/httpexpect"
)

func TestEcho(t *testing.T) {
	common.ExtraTest(httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(newHandler()),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: httpexpect.NewAssertReporter(t),
	}))
}
