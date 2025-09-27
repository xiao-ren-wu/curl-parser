package curl_parser

import (
	"strings"
	"testing"
)

func TestCurlParser_Parse(t *testing.T) {
	tests := []struct {
		name                string
		curlCommand         string
		wantMethod          string
		wantURL             string
		wantBaseURL         string
		wantPath            string
		wantHeaders         map[string]string
		wantBody            string
		wantQuery           map[string]string
		wantRawCookie       string
		wantParsedCookies   map[string]string
		wantUserAgent       string
		wantAuth            string
		wantReferer         string
		wantProxy           string
		wantConnectTimeout  int
		wantMaxTime         int
		wantInsecure        bool
		wantCACert          string
		wantCookieJar       string
		wantFollowRedirects bool
		wantErr             bool
	}{
		{
			name:              "Simple GET request",
			curlCommand:       `curl https://httpbin.org/get`,
			wantMethod:        "GET",
			wantURL:           "https://httpbin.org/get",
			wantBaseURL:       "https://httpbin.org",
			wantPath:          "/get",
			wantHeaders:       map[string]string{},
			wantBody:          "",
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:        "GET request with headers",
			curlCommand: `curl -H "Content-Type: application/json" -H "Authorization: Bearer token" https://httpbin.org/get`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/get",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/get",
			wantHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
			wantBody:          "",
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:        "POST request with data",
			curlCommand: `curl -X POST -d '{"key":"value"}' -H "Content-Type: application/json" https://httpbin.org/post`,
			wantMethod:  "POST",
			wantURL:     "https://httpbin.org/post",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/post",
			wantHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantBody:          `{"key":"value"}`,
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:              "POST request with form data",
			curlCommand:       `curl -F "key1=value1" -F "key2=value2" https://httpbin.org/post`,
			wantMethod:        "POST",
			wantURL:           "https://httpbin.org/post",
			wantBaseURL:       "https://httpbin.org",
			wantPath:          "/post",
			wantHeaders:       map[string]string{},
			wantBody:          "key1=value1&key2=value2",
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:        "GET request with query parameters",
			curlCommand: `curl https://httpbin.org/get?param1=value1&param2=value2`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/get?param1=value1&param2=value2",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/get",
			wantHeaders: map[string]string{},
			wantBody:    "",
			wantQuery: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:              "PUT request with explicit method",
			curlCommand:       `curl -X PUT --data "key=value" https://httpbin.org/put`,
			wantMethod:        "PUT",
			wantURL:           "https://httpbin.org/put",
			wantBaseURL:       "https://httpbin.org",
			wantPath:          "/put",
			wantHeaders:       map[string]string{},
			wantBody:          "key=value",
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:              "DELETE request",
			curlCommand:       `curl -X DELETE https://httpbin.org/delete`,
			wantMethod:        "DELETE",
			wantURL:           "https://httpbin.org/delete",
			wantBaseURL:       "https://httpbin.org",
			wantPath:          "/delete",
			wantHeaders:       map[string]string{},
			wantBody:          "",
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:        "Request with cookies",
			curlCommand: `curl -H "Cookie: name1=value1; name2=value2" https://httpbin.org/cookies`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/cookies",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/cookies",
			wantHeaders: map[string]string{
				"Cookie": "name1=value1; name2=value2",
			},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "name1=value1; name2=value2",
			wantParsedCookies: map[string]string{
				"name1": "value1",
				"name2": "value2",
			},
			wantErr: false,
		},
		{
			name:        "Multiline curl command",
			curlCommand: "curl -X POST \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"key\":\"value\"}' \\\n  https://httpbin.org/post",
			wantMethod:  "POST",
			wantURL:     "https://httpbin.org/post",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/post",
			wantHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			wantBody:          `{"key":"value"}`,
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:              "Request with data-raw",
			curlCommand:       `curl --request POST --url https://httpbin.org/post --data-raw '{"key": "value"}'`,
			wantMethod:        "POST",
			wantURL:           "https://httpbin.org/post",
			wantBaseURL:       "https://httpbin.org",
			wantPath:          "/post",
			wantHeaders:       map[string]string{},
			wantBody:          `{"key": "value"}`,
			wantQuery:         map[string]string{},
			wantRawCookie:     "",
			wantParsedCookies: map[string]string{},
			wantErr:           false,
		},
		{
			name:        "Request with multiple cookies",
			curlCommand: `curl -H "Cookie: sessionId=abc123; userId=456; theme=dark" https://httpbin.org/cookies`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/cookies",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/cookies",
			wantHeaders: map[string]string{
				"Cookie": "sessionId=abc123; userId=456; theme=dark",
			},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "sessionId=abc123; userId=456; theme=dark",
			wantParsedCookies: map[string]string{
				"sessionId": "abc123",
				"userId":    "456",
				"theme":     "dark",
			},
			wantErr: false,
		},
		{
			name:        "Request with cookies containing special characters",
			curlCommand: `curl -H "Cookie: name=value%20with%20spaces; encoded=test%2Bdata" https://httpbin.org/cookies`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/cookies",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/cookies",
			wantHeaders: map[string]string{
				"Cookie": "name=value%20with%20spaces; encoded=test%2Bdata",
			},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "name=value%20with%20spaces; encoded=test%2Bdata",
			wantParsedCookies: map[string]string{
				"name":    "value%20with%20spaces",
				"encoded": "test%2Bdata",
			},
			wantErr: false,
		},
		{
			name:        "Request with single cookie",
			curlCommand: `curl -H "Cookie: single=value" https://httpbin.org/cookies`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/cookies",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/cookies",
			wantHeaders: map[string]string{
				"Cookie": "single=value",
			},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "single=value",
			wantParsedCookies: map[string]string{
				"single": "value",
			},
			wantErr: false,
		},
		{
			name:          "Request with -b cookie parameter (single quotes)",
			curlCommand:   `curl -b 'sessionId=abc123; userId=456' https://httpbin.org/cookies`,
			wantMethod:    "GET",
			wantURL:       "https://httpbin.org/cookies",
			wantBaseURL:   "https://httpbin.org",
			wantPath:      "/cookies",
			wantHeaders:   map[string]string{},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "sessionId=abc123; userId=456",
			wantParsedCookies: map[string]string{
				"sessionId": "abc123",
				"userId":    "456",
			},
			wantErr: false,
		},
		{
			name:          "Request with -b cookie parameter (double quotes)",
			curlCommand:   `curl -b "name=value; theme=dark" https://httpbin.org/cookies`,
			wantMethod:    "GET",
			wantURL:       "https://httpbin.org/cookies",
			wantBaseURL:   "https://httpbin.org",
			wantPath:      "/cookies",
			wantHeaders:   map[string]string{},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "name=value; theme=dark",
			wantParsedCookies: map[string]string{
				"name":  "value",
				"theme": "dark",
			},
			wantErr: false,
		},
		{
			name:          "Request with --cookie parameter",
			curlCommand:   `curl --cookie "sessionId=abc123; userId=456; theme=light" https://httpbin.org/cookies`,
			wantMethod:    "GET",
			wantURL:       "https://httpbin.org/cookies",
			wantBaseURL:   "https://httpbin.org",
			wantPath:      "/cookies",
			wantHeaders:   map[string]string{},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "sessionId=abc123; userId=456; theme=light",
			wantParsedCookies: map[string]string{
				"sessionId": "abc123",
				"userId":    "456",
				"theme":     "light",
			},
			wantErr: false,
		},
		{
			name:          "Request with -b cookie parameter (no quotes)",
			curlCommand:   `curl -b sessionId=abc123;userId=456 https://httpbin.org/cookies`,
			wantMethod:    "GET",
			wantURL:       "https://httpbin.org/cookies",
			wantBaseURL:   "https://httpbin.org",
			wantPath:      "/cookies",
			wantHeaders:   map[string]string{},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "sessionId=abc123;userId=456",
			wantParsedCookies: map[string]string{
				"sessionId": "abc123",
				"userId":    "456",
			},
			wantErr: false,
		},
		{
			name:        "Request with both -b and -H Cookie (should prefer -b)",
			curlCommand: `curl -b "preferred=value1" -H "Cookie: ignored=value2" https://httpbin.org/cookies`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/cookies",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/cookies",
			wantHeaders: map[string]string{
				"Cookie": "ignored=value2",
			},
			wantBody:      "",
			wantQuery:     map[string]string{},
			wantRawCookie: "preferred=value1",
			wantParsedCookies: map[string]string{
				"preferred": "value1",
			},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with User-Agent",
			curlCommand:         `curl -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" https://httpbin.org/get`,
			wantMethod:          "GET",
			wantURL:             "https://httpbin.org/get",
			wantBaseURL:         "https://httpbin.org",
			wantPath:            "/get",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with authentication",
			curlCommand:         `curl -u "admin:secret123" https://httpbin.org/basic-auth/admin/secret123`,
			wantMethod:          "GET",
			wantURL:             "https://httpbin.org/basic-auth/admin/secret123",
			wantBaseURL:         "https://httpbin.org",
			wantPath:            "/basic-auth/admin/secret123",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "admin:secret123",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with referer",
			curlCommand:         `curl --referer "https://google.com" https://api.example.com/headers`,
			wantMethod:          "GET",
			wantURL:             "https://api.example.com/headers",
			wantBaseURL:         "https://api.example.com",
			wantPath:            "/headers",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "https://google.com",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with proxy",
			curlCommand:         `curl --proxy "http://proxy.example.com:8080" https://api.example.com/ip`,
			wantMethod:          "GET",
			wantURL:             "https://api.example.com/ip",
			wantBaseURL:         "https://api.example.com",
			wantPath:            "/ip",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "http://proxy.example.com:8080",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with timeouts",
			curlCommand:         `curl --connect-timeout 30 --max-time 60 https://httpbin.org/delay/5`,
			wantMethod:          "GET",
			wantURL:             "https://httpbin.org/delay/5",
			wantBaseURL:         "https://httpbin.org",
			wantPath:            "/delay/5",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  30,
			wantMaxTime:         60,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with SSL options",
			curlCommand:         `curl --insecure --cacert /path/to/ca-cert.pem https://self-signed.example.com`,
			wantMethod:          "GET",
			wantURL:             "https://self-signed.example.com",
			wantBaseURL:         "https://self-signed.example.com",
			wantPath:            "",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        true,
			wantCACert:          "/path/to/ca-cert.pem",
			wantCookieJar:       "",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with cookie jar",
			curlCommand:         `curl -c cookies.txt https://httpbin.org/cookies/set/session/abc123`,
			wantMethod:          "GET",
			wantURL:             "https://httpbin.org/cookies/set/session/abc123",
			wantBaseURL:         "https://httpbin.org",
			wantPath:            "/cookies/set/session/abc123",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "cookies.txt",
			wantFollowRedirects: false,
			wantErr:             false,
		},
		{
			name:                "Request with follow redirects",
			curlCommand:         `curl -L https://httpbin.org/redirect/3`,
			wantMethod:          "GET",
			wantURL:             "https://httpbin.org/redirect/3",
			wantBaseURL:         "https://httpbin.org",
			wantPath:            "/redirect/3",
			wantHeaders:         map[string]string{},
			wantBody:            "",
			wantQuery:           map[string]string{},
			wantRawCookie:       "",
			wantParsedCookies:   map[string]string{},
			wantUserAgent:       "",
			wantAuth:            "",
			wantReferer:         "",
			wantProxy:           "",
			wantConnectTimeout:  0,
			wantMaxTime:         0,
			wantInsecure:        false,
			wantCACert:          "",
			wantCookieJar:       "",
			wantFollowRedirects: true,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewCurlParser(tt.curlCommand)
			got, err := cp.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("CurlParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Method != tt.wantMethod {
				t.Errorf("Method = %v, want %v", got.Method, tt.wantMethod)
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %v, want %v", got.URL, tt.wantURL)
			}

			// Check BaseURL and Path
			if got.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %v, want %v", got.BaseURL, tt.wantBaseURL)
			}

			if got.Path != tt.wantPath {
				t.Errorf("Path = %v, want %v", got.Path, tt.wantPath)
			}

			// Check headers
			for key, value := range tt.wantHeaders {
				if got.Headers[key] != value {
					t.Errorf("Header %s = %v, want %v", key, got.Headers[key], value)
				}
			}

			// Check that we don't have extra headers
			for key := range got.Headers {
				if _, exists := tt.wantHeaders[key]; !exists {
					t.Errorf("Unexpected header: %s = %v", key, got.Headers[key])
				}
			}

			if got.Body != tt.wantBody {
				t.Errorf("Body = %v, want %v", got.Body, tt.wantBody)
			}

			// Check query parameters
			for key, value := range tt.wantQuery {
				if got.Query[key] != value {
					t.Errorf("Query %s = %v, want %v", key, got.Query[key], value)
				}
			}

			// Check that we don't have extra query parameters
			for key := range got.Query {
				if _, exists := tt.wantQuery[key]; !exists {
					t.Errorf("Unexpected query parameter: %s = %v", key, got.Query[key])
				}
			}

			// Check raw cookie
			if got.RawCookie != tt.wantRawCookie {
				t.Errorf("RawCookie = %v, want %v", got.RawCookie, tt.wantRawCookie)
			}

			// Check parsed cookies
			for key, value := range tt.wantParsedCookies {
				if got.ParsedCookies[key] != value {
					t.Errorf("ParsedCookie %s = %v, want %v", key, got.ParsedCookies[key], value)
				}
			}

			// Check that we don't have extra parsed cookies
			for key := range got.ParsedCookies {
				if _, exists := tt.wantParsedCookies[key]; !exists {
					t.Errorf("Unexpected parsed cookie: %s = %v", key, got.ParsedCookies[key])
				}
			}

			// Check new fields
			if got.UserAgent != tt.wantUserAgent {
				t.Errorf("UserAgent = %v, want %v", got.UserAgent, tt.wantUserAgent)
			}

			if got.Auth != tt.wantAuth {
				t.Errorf("Auth = %v, want %v", got.Auth, tt.wantAuth)
			}

			if got.Referer != tt.wantReferer {
				t.Errorf("Referer = %v, want %v", got.Referer, tt.wantReferer)
			}

			if got.Proxy != tt.wantProxy {
				t.Errorf("Proxy = %v, want %v", got.Proxy, tt.wantProxy)
			}

			if got.ConnectTimeout != tt.wantConnectTimeout {
				t.Errorf("ConnectTimeout = %v, want %v", got.ConnectTimeout, tt.wantConnectTimeout)
			}

			if got.MaxTime != tt.wantMaxTime {
				t.Errorf("MaxTime = %v, want %v", got.MaxTime, tt.wantMaxTime)
			}

			if got.Insecure != tt.wantInsecure {
				t.Errorf("Insecure = %v, want %v", got.Insecure, tt.wantInsecure)
			}

			if got.CACert != tt.wantCACert {
				t.Errorf("CACert = %v, want %v", got.CACert, tt.wantCACert)
			}

			if got.CookieJar != tt.wantCookieJar {
				t.Errorf("CookieJar = %v, want %v", got.CookieJar, tt.wantCookieJar)
			}

			if got.FollowRedirects != tt.wantFollowRedirects {
				t.Errorf("FollowRedirects = %v, want %v", got.FollowRedirects, tt.wantFollowRedirects)
			}
		})
	}
}

func TestCurlParser_extractURL(t *testing.T) {
	tests := []struct {
		name        string
		curlCommand string
		want        string
		wantErr     bool
	}{
		{
			name:        "URL with quotes",
			curlCommand: `curl -H "Content-Type: application/json" "https://httpbin.org/get"`,
			want:        "https://httpbin.org/get",
			wantErr:     false,
		},
		{
			name:        "URL without quotes",
			curlCommand: `curl -X POST https://httpbin.org/post`,
			want:        "https://httpbin.org/post",
			wantErr:     false,
		},
		{
			name:        "No URL",
			curlCommand: `curl -H "Content-Type: application/json"`,
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewCurlParser(tt.curlCommand)
			// Clean the command like in Parse method
			cmd := strings.ReplaceAll(cp.curlCommand, "\\\n", " ")
			cmd = strings.ReplaceAll(cmd, "\\", "")
			cmd = strings.TrimSpace(cmd)
			if strings.HasPrefix(cmd, "curl ") {
				cmd = strings.TrimPrefix(cmd, "curl ")
			}

			got, err := cp.extractURL(cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("CurlParser.extractURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CurlParser.extractURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurlParser_extractMethod(t *testing.T) {
	tests := []struct {
		name        string
		curlCommand string
		want        string
	}{
		{
			name:        "Explicit GET method",
			curlCommand: `curl -X GET https://httpbin.org/get`,
			want:        "GET",
		},
		{
			name:        "Explicit POST method",
			curlCommand: `curl -X POST https://httpbin.org/post`,
			want:        "POST",
		},
		{
			name:        "Long form request method",
			curlCommand: `curl --request PUT https://httpbin.org/put`,
			want:        "PUT",
		},
		{
			name:        "Implicit POST with data",
			curlCommand: `curl -d "key=value" https://httpbin.org/post`,
			want:        "POST",
		},
		{
			name:        "Implicit POST with form",
			curlCommand: `curl -F "key=value" https://httpbin.org/post`,
			want:        "POST",
		},
		{
			name:        "Default GET",
			curlCommand: `curl https://httpbin.org/get`,
			want:        "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewCurlParser(tt.curlCommand)
			// Clean the command like in Parse method
			cmd := strings.ReplaceAll(cp.curlCommand, "\\\n", " ")
			cmd = strings.ReplaceAll(cmd, "\\", "")
			cmd = strings.TrimSpace(cmd)
			if strings.HasPrefix(cmd, "curl ") {
				cmd = strings.TrimPrefix(cmd, "curl ")
			}

			if got := cp.extractMethod(cmd); got != tt.want {
				t.Errorf("CurlParser.extractMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurlParser_extractBody(t *testing.T) {
	tests := []struct {
		name        string
		curlCommand string
		want        string
	}{
		{
			name:        "Data with single quotes",
			curlCommand: `curl -d '{"key":"value"}' https://httpbin.org/post`,
			want:        `{"key":"value"}`,
		},
		{
			name:        "Data without quotes",
			curlCommand: `curl -d key=value https://httpbin.org/post`,
			want:        `key=value`,
		},
		{
			name:        "Form data",
			curlCommand: `curl -F "key1=value1" -F "key2=value2" https://httpbin.org/post`,
			want:        `key1=value1&key2=value2`,
		},
		{
			name:        "No body",
			curlCommand: `curl https://httpbin.org/get`,
			want:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewCurlParser(tt.curlCommand)
			// Clean the command like in Parse method
			cmd := strings.ReplaceAll(cp.curlCommand, "\\\n", " ")
			cmd = strings.ReplaceAll(cmd, "\\", "")
			cmd = strings.TrimSpace(cmd)
			if strings.HasPrefix(cmd, "curl ") {
				cmd = strings.TrimPrefix(cmd, "curl ")
			}

			if got := cp.extractBody(cmd); got != tt.want {
				t.Errorf("CurlParser.extractBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
