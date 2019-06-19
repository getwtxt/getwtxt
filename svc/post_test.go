package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/getwtxt/registry"
)

var apiPostUserCases = []struct {
	name    string
	nick    string
	uri     string
	wantErr bool
}{
	{
		name:    "Known Good User",
		nick:    "gbmor",
		uri:     "https://gbmor.dev/twtxt.txt",
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
	portnum := fmt.Sprintf(":%v", confObj.Port)
	twtxtCache = registry.NewIndex(nil)

	for _, tt := range apiPostUserCases {
		t.Run(tt.name, func(t *testing.T) {
			params := url.Values{}
			params.Set("url", tt.uri)
			params.Set("nickname", tt.nick)

			req, err := http.NewRequest("POST", "https://localhost"+portnum+"/api/plain/users", strings.NewReader(params.Encode()))
			if err != nil {
				t.Errorf("%v\n", err)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			apiEndpointPOSTHandler(rr, req)

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
func Benchmark_apiPostUser(b *testing.B) {
	initTestConf()
	portnum := fmt.Sprintf(":%v", confObj.Port)
	twtxtCache = registry.NewIndex(nil)

	params := url.Values{}
	params.Set("url", "https://gbmor.dev/twtxt.txt")
	params.Set("nickname", "gbmor")
	req, _ := http.NewRequest("POST", "https://localhost"+portnum+"/api/plain/users", strings.NewReader(params.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		apiEndpointPOSTHandler(rr, req)

		b.StopTimer()
		twtxtCache = registry.NewIndex(nil)
		b.StartTimer()
	}
}
