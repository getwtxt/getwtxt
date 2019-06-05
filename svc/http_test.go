package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_log400(t *testing.T) {
	initTestConf()
	t.Run("log400", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/400", nil)
		log400(w, req, "400 Test")
		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Didn't receive 400, received: %v\n", resp.StatusCode)
		}
	})
}

func Test_log404(t *testing.T) {
	initTestConf()
	t.Run("log404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/404", nil)
		log404(w, req, errors.New("404 Test"))
		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Didn't receive 404, received: %v\n", resp.StatusCode)
		}
	})
}

func Test_log500(t *testing.T) {
	initTestConf()
	t.Run("log500", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/500", nil)
		log500(w, req, errors.New("500 Test"))
		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Didn't receive 500, received: %v\n", resp.StatusCode)
		}
	})
}
