package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/kelseyhightower/envconfig"
)

var cfg = struct {
	Host           string
	Port           int           `default:"8080"`
	DrainTimeout   time.Duration `default:"25s"`
	K8sGracePeriod time.Duration `default:"30s"`
}{}

var static = map[string]string{
	"/":            index,
	"/index":       index,
	"/index.htm":   index,
	"/index.html":  index,
	"/keybase.txt": keybase,
	"/healthz":     "",
}

func main() {
	ctx := context.Background()
	envconfig.MustProcess("JOA", &cfg)

	// server
	r := http.HandlerFunc(root)
	handler := gziphandler.GzipHandler(r)
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	server := &http.Server{Addr: addr, Handler: handler}
	go listen(addr, server)

	// await term
	awaitTerm()

	// await graceful shutdown
	ctx, cancel := context.WithTimeout(ctx, cfg.DrainTimeout)
	awaitShutdown(ctx, server)

	// exit
	log.Println("bye bye")
	cancel()
	os.Exit(0)
}

func root(w http.ResponseWriter, r *http.Request) {
	if data, ok := static[r.URL.Path]; ok {
		if _, err := io.WriteString(w, data); err != nil {
			log.Println("couldn't send response:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func listen(addr string, server *http.Server) {
	log.Printf("server running at http://%s ...", addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalln("couldn't start server:", err.Error())
	}
}

func awaitShutdown(ctx context.Context, server *http.Server) {
	log.Println("performing graceful shutdown")

	done := make(chan bool)

	go func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Println("couldn't shutdown server:", err.Error())
		}

		done <- true
	}()

	select {
	case <-time.After(cfg.K8sGracePeriod):
		log.Println("graceful shutdown failed")
	case <-done:
	}
}

func awaitTerm() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT)
	<-sig
}
