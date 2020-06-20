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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// The first few are testing whether the landing page is
// being sent correctly. If i change the base behavior of
//    /api
//    /api/plain
// later, then I'll change the tests.

func basicHandlerTest(path, name string, t *testing.T) {
	initTestConf()

	t.Run(name, func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)

		staticHandler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}

		bt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		if !reflect.DeepEqual(bt, staticCache.index) {
			t.Errorf("Byte mismatch\n")
		}
	})
}
func Test_indexHandler(t *testing.T) {
	basicHandlerTest("http://localhost"+testport+"/", "indexHandler", t)
}
func Benchmark_indexHandler(b *testing.B) {
	initTestConf()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost"+testport+"/", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		staticHandler(w, req)
	}
}
func Test_apiBaseHandler(t *testing.T) {
	basicHandlerTest("http://localhost"+testport+"/api", "apiBaseHandler", t)
}
func Test_apiFormatHandler(t *testing.T) {
	basicHandlerTest("http://localhost"+testport+"/api/format", "apiFormatHandler", t)
}

var endpointCases = []struct {
	name   string
	req    *http.Request
	status int
}{
	{
		name:   "Regular Query: /api/plain/users",
		req:    httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/users", nil),
		status: http.StatusOK,
	},
	{
		name:   "Regular Query: /api/plain/mentions",
		req:    httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/mentions", nil),
		status: http.StatusOK,
	},
	{
		name:   "Regular Query: /api/plain/tweets",
		req:    httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/tweets", nil),
		status: http.StatusOK,
	},
	{
		name:   "Invalid Endpoint: /api/plain/statuses",
		req:    httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/statuses", nil),
		status: http.StatusNotFound,
	},
}

func Test_apiEndpointHandler(t *testing.T) {
	initTestConf()
	mockRegistry()

	for _, tt := range endpointCases {

		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			apiEndpointHandler(w, tt.req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.status {
				t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
			}

			if tt.status == http.StatusOK {
				var body []byte
				buf := bytes.NewBuffer(body)

				err := resp.Write(buf)
				if err != nil {
					t.Errorf("%v\n", err)
				}
				if buf == nil {
					t.Errorf("Got nil\n")
				}
				if len(buf.Bytes()) == 0 {
					t.Errorf("Got zero data\n")
				}
			}
		})
	}
}
func Benchmark_apiEndpointHandler(b *testing.B) {
	initTestConf()
	mockRegistry()

	w := httptest.NewRecorder()
	b.ResetTimer()

	for _, tt := range endpointCases {
		for i := 0; i < b.N; i++ {
			apiEndpointHandler(w, tt.req)
		}
	}
}

func Test_apiTagsBaseHandler(t *testing.T) {
	initTestConf()
	mockRegistry()

	t.Run("apiTagsBaseHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/tags", nil)

		apiTagsBaseHandler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}

		bd, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		if len(bd) == 0 {
			t.Errorf("Got no data from registry\n")
		}
	})
}
func Benchmark_apiTagsBaseHandler(b *testing.B) {
	initTestConf()
	mockRegistry()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/tags", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		apiTagsBaseHandler(w, r)
	}
}
func Test_apiTagsHandler(t *testing.T) {
	initTestConf()
	mockRegistry()

	t.Run("apiTagsHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/tags/programming", nil)

		apiTagsHandler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v\n", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
		if len(data) == 0 {
			t.Errorf("Got no data: %v\n", data)
		}
	})
}

func Benchmark_apiTagsHandler(b *testing.B) {
	initTestConf()
	mockRegistry()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost"+testport+"/api/plain/tags/programming", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		apiTagsHandler(w, r)
	}
}

func Test_cssHandler(t *testing.T) {
	initTestConf()

	name := "CSS Handler Test"
	css, err := ioutil.ReadFile("../assets/style.css")
	if err != nil {
		t.Errorf("Couldn't read ../assets/style.css: %v\n", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost"+testport+"/css", nil)

	t.Run(name, func(t *testing.T) {
		staticHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			t.Errorf("cssHandler(): %v\n", resp.StatusCode)
		}
		if !reflect.DeepEqual(body, css) {
			t.Errorf("cssHandler(): Byte mismatch\n")
		}
	})
}
