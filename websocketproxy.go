package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

func WebsocketProxy(target *url.URL) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, err := net.Dial("tcp", target.Host)
		if err != nil {
			TemplateHandler(http.StatusBadGateway, "bad_gateway.html").ServeHTTP(w, r)
			log.Printf("error dialing websocket backend %s: %v", target, err)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Not a hijacker?", 500)
			return
		}
		nc, _, err := hj.Hijack()
		if err != nil {
			log.Printf("hijack error: %v", err)
			return
		}
		defer nc.Close()
		defer d.Close()
		err = r.Write(d)
		if err != nil {
			log.Printf("error copying request to target: %v", err)
			return
		}
		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}
		go cp(d, nc)
		go cp(nc, d)
		<-errc
	})
}
