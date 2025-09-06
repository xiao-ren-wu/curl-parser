package curl_parser

import (
	"strings"
	"testing"
)

func TestCurlParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		curlCommand string
		wantMethod  string
		wantURL     string
		wantBaseURL string
		wantPath    string
		wantHeaders map[string]string
		wantBody    string
		wantQuery   map[string]string
		wantErr     bool
	}{
		{
			name:        "Simple GET request",
			curlCommand: `curl https://httpbin.org/get`,
			wantMethod:  "GET",
			wantURL:     "https://httpbin.org/get",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/get",
			wantHeaders: map[string]string{},
			wantBody:    "",
			wantQuery:   map[string]string{},
			wantErr:     false,
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
			wantBody:  "",
			wantQuery: map[string]string{},
			wantErr:   false,
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
			wantBody:  `{"key":"value"}`,
			wantQuery: map[string]string{},
			wantErr:   false,
		},
		{
			name:        "POST request with form data",
			curlCommand: `curl -F "key1=value1" -F "key2=value2" https://httpbin.org/post`,
			wantMethod:  "POST",
			wantURL:     "https://httpbin.org/post",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/post",
			wantHeaders: map[string]string{},
			wantBody:    "key1=value1&key2=value2",
			wantQuery:   map[string]string{},
			wantErr:     false,
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
			wantErr: false,
		},
		{
			name:        "PUT request with explicit method",
			curlCommand: `curl -X PUT --data "key=value" https://httpbin.org/put`,
			wantMethod:  "PUT",
			wantURL:     "https://httpbin.org/put",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/put",
			wantHeaders: map[string]string{},
			wantBody:    "key=value",
			wantQuery:   map[string]string{},
			wantErr:     false,
		},
		{
			name:        "DELETE request",
			curlCommand: `curl -X DELETE https://httpbin.org/delete`,
			wantMethod:  "DELETE",
			wantURL:     "https://httpbin.org/delete",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/delete",
			wantHeaders: map[string]string{},
			wantBody:    "",
			wantQuery:   map[string]string{},
			wantErr:     false,
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
			wantBody:  "",
			wantQuery: map[string]string{},
			wantErr:   false,
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
			wantBody:  `{"key":"value"}`,
			wantQuery: map[string]string{},
			wantErr:   false,
		},
		{
			name:        "Request with data-raw",
			curlCommand: `curl --request POST --url https://httpbin.org/post --data-raw '{"key": "value"}'`,
			wantMethod:  "POST",
			wantURL:     "https://httpbin.org/post",
			wantBaseURL: "https://httpbin.org",
			wantPath:    "/post",
			wantHeaders: map[string]string{},
			wantBody:    `{"key": "value"}`,
			wantQuery:   map[string]string{},
			wantErr:     false,
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
