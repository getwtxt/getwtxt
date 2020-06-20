/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "git.sr.ht/~gbmor/getwtxt/svc"

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"git.sr.ht/~gbmor/getwtxt/registry"
)

var apiPostUserCases = []struct {
	name    string
	nick    string
	uri     string
	wantErr bool
}{
	{
		name:    "Known Good User",
		nick:    "getwtxttest",
		uri:     testTwtxtURL,
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
	twtxtCache = registry.New(nil)

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
	twtxtCache = registry.New(nil)

	params := url.Values{}
	params.Set("url", testTwtxtURL)
	params.Set("nickname", "gbmor")
	req, _ := http.NewRequest("POST", "https://localhost"+portnum+"/api/plain/users", strings.NewReader(params.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		apiEndpointPOSTHandler(rr, req)

		b.StopTimer()
		twtxtCache = registry.New(nil)
		b.StartTimer()
	}
}
