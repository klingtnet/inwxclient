package inwxclient

import (
	"net/http"
	"net/http/cookiejar"
)

func newClient(baseClient *http.Client) (*http.Client, error) {
	var baseTransport http.RoundTripper
	if baseClient.Transport != nil {
		baseTransport = baseClient.Transport
	} else {
		baseTransport = http.DefaultTransport
	}
	baseClient.Transport = newCustomTransport(baseTransport)

	if baseClient.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		baseClient.Jar = jar
	}

	return baseClient, nil
}

type customTransport struct {
	baseRoundTripper http.RoundTripper
}

func newCustomTransport(baseTransport http.RoundTripper) *customTransport {
	return &customTransport{
		baseRoundTripper: baseTransport,
	}
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "github.com/klingtnet/inwxclient")
	req.Header.Set("Content-Type", "application/json")

	return c.baseRoundTripper.RoundTrip(req)
}
