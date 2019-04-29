package github

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

const (
	testToken = "810a5bdaafc6dd30b1d9979215935871"
)

type expected struct {
	ghUrl string
	token string
	pkg   packageQuery
}

func TestReadUrl(t *testing.T) {
	cases := map[string]struct {
		expected
		skip bool
		err  string
		url  string
	}{
		"should parse simple url": {
			skip: false,
			url:  "github://github.com/foo/bar",
			expected: expected{
				pkg: packageQuery{
					owner:    "foo",
					repo:     "bar",
					version:  "latest",
					location: "github.com/foo/bar/latest",
				},
			},
		},
		"should fail if URL has no host": {
			skip: false,
			url:  "github:///",
			err:  errNoHost.Error(),
		},
		"should fail if URL format is invalid": {
			skip: false,
			url:  "github://github.com/foo",
			err:  errBadUrl.Error(),
		},
		"should parse package version and token": {
			skip: false,
			url:  "github://github.com/foo/bar?version=v1.0&token=" + testToken,
			expected: expected{
				token: testToken,
				pkg: packageQuery{
					owner:    "foo",
					repo:     "bar",
					version:  "v1.0",
					location: "github.com/foo/bar/v1.0",
				},
			},
		},
		"should parse enterprise GitHub URL": {
			skip: false,
			url:  "github://github.example.com:8888/foo/bar?version=v1.0&token=" + testToken,
			expected: expected{
				ghUrl: "https://github.example.com:8888/",
				token: testToken,
				pkg: packageQuery{
					owner:    "foo",
					repo:     "bar",
					version:  "v1.0",
					location: "github.example.com/foo/bar/v1.0",
				},
			},
		},
		"should parse enterprise GitHub URL with custom protocol and path": {
			skip: false,
			url:  "github://github.example.com/service/foo/bar?version=v1.0&protocol=http&token=" + testToken,
			expected: expected{
				ghUrl: "http://github.example.com/service/",
				token: testToken,
				pkg: packageQuery{
					owner:    "foo",
					repo:     "bar",
					version:  "v1.0",
					location: "github.example.com/service/foo/bar/v1.0",
				},
			},
		},
		"should parse simple GH enterprise url": {
			skip: false,
			url:  "github://github.example.com/foo/bar",
			expected: expected{
				ghUrl: "https://github.example.com/",
				pkg: packageQuery{
					owner:    "foo",
					repo:     "bar",
					version:  "latest",
					location: "github.example.com/foo/bar/latest",
				},
			},
		},
	}

	ctx := context.Background()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			if c.skip {
				t.SkipNow()
				return
			}
			defer func() {
				if r := recover(); r != nil {
					t.Fatal(r)
				}
			}()
			uri, err := url.Parse(c.url)
			if err != nil {
				t.Fatal(err)
			}

			dc, err := readUrl(ctx, uri)
			if c.err != "" {
				assert.EqualError(t, err, c.err)
				return
			}

			assert.NoError(t, err)
			if c.expected.ghUrl != "" {
				gotUrl := dc.ghClient.BaseURL.String()
				assert.Equal(t, c.expected.ghUrl, gotUrl)
			}

			// token check
			if c.expected.token != "" {
				// dirty hack to extract http client from gh client
				//clientPtr := reflect.ValueOf(dc.ghClient)
				//r := reflect.Indirect(clientPtr)
				//cf := r.FieldByName("client")
				//fieldPtr := (**http.Client)(unsafe.Pointer(cf.UnsafeAddr()))

				httpClient := dc.httpClient
				if httpClient.Transport == nil {
					t.Fatalf("token storage is nil but token was expected")
				}

				ts, ok := httpClient.Transport.(*oauth2.Transport)
				if !ok || ts == nil || ts.Source == nil {
					t.Fatal("token storage was not found but token was expected")
				}

				tkn, err := ts.Source.Token()
				assert.NoError(t, err)

				assert.Equal(t, c.expected.token, tkn.AccessToken)
			}

			assert.Equal(t, c.expected.pkg, dc.pkg)
		})
	}
}
