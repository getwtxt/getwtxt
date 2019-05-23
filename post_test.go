package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var apiPostUserCases = []struct {
	name    string
	nick    string
	uri     string
	wantErr bool
}{
	{
		name:    "Known Good User",
		nick:    "soltempore",
		uri:     "https://enotty.dk/soltempore.txt",
		wantErr: false,
	},
	{
		name:    "Missing URI",
		nick:    "missinguri",
		uri:     "",
		wantErr: true,
	},
	{
		name:    "Missing Nickname",
		nick:    "",
		uri:     "https://example.com/twtxt.txt",
		wantErr: true,
	},
	{
		name:    "Missing URI and Nickname",
		nick:    "",
		uri:     "",
		wantErr: true,
	},
}

func Test_apiPostUser(t *testing.T) {
	initTestConf()

	for _, tt := range apiPostUserCases {
		t.Run(tt.name, func(t *testing.T) {
			params := url.Values{}
			params.Set("url", tt.uri)
			params.Set("nickname", tt.nick)

			req, err := http.NewRequest("POST", "http://localhost"+testport+"/api/plain/users", strings.NewReader(params.Encode()))
			if err != nil {
				t.Errorf("%v\n", err)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(apiEndpointPOSTHandler)

			handler.ServeHTTP(rr, req)

			if !tt.wantErr {
				if rr.Code != http.StatusOK {
					t.Errorf("Received unexpected non-200 response: %v\n", rr.Code)
				}
			} else {
				if rr.Code != http.StatusBadRequest {
					t.Errorf("Expected 400 Bad Request, but received: %v\n", rr.Code)
				}
			}
		})
	}
}
