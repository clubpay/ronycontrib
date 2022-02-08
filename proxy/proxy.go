package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxy struct {
	routes map[string]*httputil.ReverseProxy
}

func New() *ReverseProxy {
	return &ReverseProxy{
		routes: map[string]*httputil.ReverseProxy{},
	}
}

func (rp *ReverseProxy) AddRoute(path, targetBaseURL string) *ReverseProxy {
	rawURL := fmt.Sprintf("%s/%s",
		strings.TrimRight(targetBaseURL, "/"),
		strings.TrimLeft(path, "/"),
	)

	targetURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	rp.routes[path] = httputil.NewSingleHostReverseProxy(targetURL)

	return rp
}

func (rp *ReverseProxy) Run(port int) {
	for path, p := range rp.routes {
		http.HandleFunc(path, p.ServeHTTP)
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}
