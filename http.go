package main

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type HTTPRouter struct {
	mtx   sync.RWMutex
	hosts map[string]*url.URL
}

func NewHTTPRouter() *HTTPRouter {
	hr := HTTPRouter{
		hosts: make(map[string]*url.URL),
	}
	return &hr
}

func (hr *HTTPRouter) SetHost(host string, target *url.URL) {
	hr.mtx.Lock()
	defer hr.mtx.Unlock()
	hr.hosts[host] = target
}

func (hr *HTTPRouter) RemoveHost(host string) {
	hr.mtx.Lock()
	defer hr.mtx.Unlock()
	delete(hr.hosts, host)
}

func (hr *HTTPRouter) isWebsocket(req *http.Request) bool {
	conn_hdr := ""
	conn_hdrs := req.Header["Connection"]
	if len(conn_hdrs) > 0 {
		conn_hdr = conn_hdrs[0]
	}
	do := false
	if strings.ToLower(conn_hdr) == "upgrade" {
		upgrade_hdrs := req.Header["Upgrade"]
		if len(upgrade_hdrs) > 0 {
			do = (strings.ToLower(upgrade_hdrs[0]) == "websocket")
		}
	}
	return do
}

func (hr *HTTPRouter) makeHandlerForReq(req *http.Request, target *url.URL) http.Handler {
	if hr.isWebsocket(req) {
		return WebsocketProxy(target)
	}
	return NewSingleHostReverseProxy(target)
}

func (hr *HTTPRouter) findHandlerForReq(req *http.Request) http.Handler {
	host := strings.ToLower(req.Host)
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}
	hr.mtx.RLock()
	defer hr.mtx.RUnlock()
	if target, ok := hr.hosts[host]; ok {
		return hr.makeHandlerForReq(req, target)
	}
	// handle wildcard domains up to 2 subdomains deep
	d := strings.SplitN(host, ".", 2)
	for i := len(d); i > 0; i-- {
		if target, ok := hr.hosts["*."+strings.Join(d[len(d)-i:], ".")]; ok {
			return hr.makeHandlerForReq(req, target)
		}
	}
	return nil
}

func (hr *HTTPRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler
	handler = hr.findHandlerForReq(req)
	if handler == nil {
		handler = TemplateHandler(http.StatusNotFound, "not_found.html")
	}
	handler.ServeHTTP(w, req)
}
