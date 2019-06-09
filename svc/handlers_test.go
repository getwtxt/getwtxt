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
func Test_apiEndpointHandler(t *testing.T) {
	initTestConf()
	t.Run("apiEndpointHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api/plain/users", nil)
		apiEndpointHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
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
