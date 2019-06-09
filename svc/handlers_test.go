package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Currently, these only test for a 200 status code.
// More in-depth unit tests are planned, however, several
// of these will quickly turn into integration tests as
// they'll need more than a barebones test environment to
// get any real information. The HTTP responses are being
// tested by me by hand, mostly.

func Test_indexHandler(t *testing.T) {
	initTestConf()
	t.Run("indexHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/", nil)
		indexHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}
func Test_apiBaseHandler(t *testing.T) {
	initTestConf()
	t.Run("apiBaseHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api", nil)
		apiBaseHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}
func Test_apiFormatHandler(t *testing.T) {
	initTestConf()
	t.Run("apiFormatHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api/plain", nil)
		apiFormatHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
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
	t.Run("apiTagsBaseHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api/plain/tags", nil)
		apiTagsBaseHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}
func Test_apiTagsHandler(t *testing.T) {
	initTestConf()
	t.Run("apiTagsHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api/plain/tags/tag", nil)
		apiTagsHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}

func Test_cssHandler(t *testing.T) {
	initTestConf()

	name := "CSS Handler Test"
	css, err := ioutil.ReadFile("../assets/style.css")
	if err != nil {
		t.Errorf("Couldn't read ../assets/style.css: %v\n", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "localhost"+testport+"/css", nil)

	t.Run(name, func(t *testing.T) {
		cssHandler(w, req)
		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			t.Errorf("cssHandler(): %v\n", resp.StatusCode)
		}
		if !bytes.Equal(body, css) {
			t.Errorf("cssHandler(): Byte mismatch\n")
		}
	})
}
