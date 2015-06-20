package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/flynn/flynn/pkg/tlsconfig"
	"golang.org/x/net/context"
)

func main() {
	httpRouter := NewHTTPRouter()
	errc := make(chan error)

	ds, err := NewKubernetesDataStore()
	if err != nil {
		log.Fatal(err)
	}

	syncCtx := context.Background()
	syncCtx, stopSync := context.WithCancel(syncCtx)

	// watch kubernetes
	log.Println("starting kubernetes watcher...")
	go ds.Watch(syncCtx, httpRouter)

	if err := httpListenAndServe(httpRouter, errc); err != nil {
		stopSync()
		log.Fatal(fmt.Sprintf("error: http server: %s", err))
	}

	if err := httpsListenAndServe(httpRouter, errc); err != nil {
		stopSync()
		log.Fatal(fmt.Sprintf("error: https server: %s", err))
	}

	if err := <-errc; err != nil {
		log.Fatal(fmt.Sprintf("serve error: %s", err))
	}

	log.Println("exited")
}

func httpListenAndServe(handler http.Handler, errc chan error) error {
	listener, err := net.Listen("tcp4", ":80")
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: listener.Addr().String(),
		Handler: fwdProtoHandler{
			Handler: handler,
			Proto:   "http",
			Port:    mustPortFromAddr(listener.Addr().String()),
		},
	}
	go func() {
		errc <- server.Serve(listener)
	}()
	return nil
}

func httpsListenAndServe(handler http.Handler, errc chan error) error {
	tlsKeyPair, err := tls.LoadX509KeyPair("/etc/router/tls.crt", "/etc/router/tls.key")
	if err != nil {
		return err
	}
	certForHandshake := func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return nil, nil
	}
	tlsConfig := tlsconfig.SecureCiphers(&tls.Config{
		GetCertificate: certForHandshake,
		Certificates:   []tls.Certificate{tlsKeyPair},
	})
	listener, err := net.Listen("tcp4", ":443")
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(listener, tlsConfig)
	server := &http.Server{
		Addr: tlsListener.Addr().String(),
		Handler: fwdProtoHandler{
			Handler: handler,
			Proto:   "https",
			Port:    mustPortFromAddr(tlsListener.Addr().String()),
		},
	}
	go func() {
		errc <- server.Serve(tlsListener)
	}()
	return nil
}

func mustPortFromAddr(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	return port
}
