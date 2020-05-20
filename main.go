package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	config := getConfig()
	exporter := NewExporter(config.Metrics)

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(exporter)

	h := promhttp.HandlerFor(registry,
		promhttp.HandlerOpts{
			ErrorLog:      log.New(os.Stderr, "", log.LstdFlags),
			ErrorHandling: promhttp.ContinueOnError,
		})

	root := "/" + strings.Trim(config.Path,"/")
	http.Handle(root, h)

	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)
	log.Printf("server is running at http://%v%v", addr, root)
	log.Fatal(http.ListenAndServe(addr, nil))
}
