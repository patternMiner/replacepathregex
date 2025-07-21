package replacepathregex_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"testing"

	replacepathregex "github.com/patternMiner/replacepathregex"
)

func Testreplacepathregex(t *testing.T) {

	tests := []struct {
		name        string
		regex       string
		requestURL  string
		replacement string
		expected    string
	}{
		{
			name:        "test0",
			requestURL:  "https://example.test.com/admin/v0/blah",
			regex:       `^(.*)/admin/(.*)$`,
			replacement: "/$2",
			expected:    "/v0/blah",
		},
		{
			name:        "test1",
			requestURL:  "https://example.test.com/robots.txt",
			regex:       `https://(.*).test.com/robots\.txt$`,
			replacement: "/robots_txt/$1/robots.txt",
			expected:    "/robots_txt/example/robots.txt",
		},
		{
			name:        "test2",
			requestURL:  "https://example.test.com/home.xml",
			regex:       `https://(.*).test.com/(.*)\.xml$`,
			replacement: "/sitemap/$1/$2.xml",
			expected:    "/sitemap/example/home.xml",
		},
		{
			name:        "sitemap pinzap.shop",
			requestURL:  "https://xxxx.pinzap.shop/xx.xml",
			regex:       `https://([^/]+)/([^/]+\.xml)$`,
			replacement: "/sitemap/$1/$2",
			expected:    "/sitemap/xxxx.pinzap.shop/xx.xml",
		},
		{
			name:        "sitemap pinkoi.com",
			requestURL:  "https://pinzap.pinkoi.com/xx.xml",
			regex:       `https://([^/]+)/([^/]+\.xml)$`,
			replacement: "/sitemap/$1/$2",
			expected:    "/sitemap/pinzap.pinkoi.com/xx.xml",
		},
		{
			name:        "robots.txt pinkoi.com",
			requestURL:  "https://pinzap.pinkoi.com/robots.txt",
			regex:       `https://([^/]+)/robots\.txt$`,
			replacement: "/robots_txt/$1/robots.txt",
			expected:    "/robots_txt/pinzap.pinkoi.com/robots.txt",
		},
		{
			name:        "robots.txt pinkoi.com",
			requestURL:  "https://pinzap.pinkoi.com/robots.txt",
			regex:       `https://([^/]+)/robots\.txt$`,
			replacement: "/robots_txt/$1/robots.txt$2",
			expected:    "/robots_txt/pinzap.pinkoi.com/robots.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler, err := replacepathregex.New(context.Background(), next, &replacepathregex.Config{
				Regex:       test.regex,
				Replacement: test.replacement,
			}, "replacepathregex")

			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			method := http.MethodGet

			req := httptest.NewRequest(method, test.requestURL, nil)

			handler.ServeHTTP(recorder, req)

			if req.URL.Path != test.expected {
				t.Errorf("expected %s, got %s", test.expected, req.URL.Path)
			}
		})
	}
}
