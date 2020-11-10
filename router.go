package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type requestHandler func(context.Context, http.ResponseWriter) error

func (h requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h(ctx, w); err != nil {
		errMsg := fmt.Sprintf("An error occured: %s", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
}

func returnJSON(w http.ResponseWriter, item interface{}) error {
	resp, err := json.Marshal(item)
	if err != nil {
		return errors.Wrap(err, "Error converting response to JSON")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return nil
}

type Mux struct {
	initOnce sync.Once
	mux      *http.ServeMux
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *Mux) init() {
	m.initOnce.Do(func() {
		m.mux = http.NewServeMux()
	})
}

func (m *Mux) GET(pattern string, handler http.Handler) {
	m.init()
	h := handler
	h = AllowedMethod(h, http.MethodGet)
	h = RequestWithTimeout(h, 10*time.Second)
	m.mux.Handle(pattern, h)
}

func (m *Mux) OK(pattern string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pattern {
			http.NotFound(w, r)
			return
		}
	}
	m.GET(pattern, http.HandlerFunc(handler))
}

func (m *Mux) File(pattern string, desc string, path string) {
	handler := func(ctx context.Context, w http.ResponseWriter) error {
		output, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Error fetching %s", desc)
		}
		w.Write(output)
		return nil
	}
	m.GET(pattern, requestHandler(handler))
}

func (m *Mux) Command(pattern string, desc string, cmd string, args ...string) {
	handler := func(ctx context.Context, w http.ResponseWriter) error {
		cmd := exec.CommandContext(ctx, cmd, args...)
		output, err := cmd.Output()
		if err != nil {
			return errors.Wrapf(err, "Error fetching %s", desc)
		}
		w.Write(output)
		return nil
	}
	m.GET(pattern, requestHandler(handler))
}

func AllowedMethod(h http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func RequestWithTimeout(h http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
