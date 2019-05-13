package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testport = fmt.Sprintf(":%v", confObj.port)

func initTestConf() {
	initConfig()
	logToNull()
}

func logToNull() {
	hush, err := os.Open("/dev/null")
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.SetOutput(hush)
}

// these will be expanded later. currently, they only
// test for a 200 status code.
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
		indexHandler(w, req)
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
		indexHandler(w, req)
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
		indexHandler(w, req)
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
		indexHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}
func Test_apiTagsHandler(t *testing.T) {
	initTestConf()
	t.Run("indexHandler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "localhost"+testport+"/api/plain/tags/tag", nil)
		indexHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf(fmt.Sprintf("%v", resp.StatusCode))
		}
	})
}
func Test_cssHandler(t *testing.T) {
	initTestConf()

	name := "CSS Handler Test"
	css, err := ioutil.ReadFile("assets/style.css")
	if err != nil {
		t.Errorf("Couldn't read assets/style.css: %v\n", err)
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
