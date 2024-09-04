package yevna_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/goccy/go-json"

	. "github.com/onsi/ginkgo/v2"
)

var (
	svc       *httptest.Server
	ipInfoMap = map[string]any{
		"ip":       "1.1.1.1",
		"hostname": "one.one.one.one",
		"anycast":  true,
		"city":     "The Rocks",
		"region":   "New South Wales",
		"country":  "AU",
		"loc":      "-33.8592,151.2081",
		"org":      "AS13335 Cloudflare, Inc.",
		"postal":   "2000",
		"timezone": "Australia/Sydney",
	}
)

var _ = BeforeSuite(func() {
	g := http.NewServeMux()
	g.HandleFunc("/ipinfo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		buf, err := json.Marshal(ipInfoMap)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(buf)
	})
	svc = httptest.NewServer(g)
})
