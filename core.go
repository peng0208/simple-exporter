package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

type exporter struct {
	metrics map[string]interface{}
	sources map[string]string
	mutex   *sync.Mutex
	wg      *sync.WaitGroup
}

const (
	Gauge       = "gauge"
	Counter     = "counter"
	Summary     = "summary"
	Histogram   = "histogram"
	DefaultType = Gauge
)

func NewExporter(config MetricsConfig) *exporter {
	e := &exporter{
		make(map[string]interface{}),
		make(map[string]string),
		&sync.Mutex{},
		&sync.WaitGroup{},
	}

	for _, m := range config {
		e.registerMetric(m)
	}
	return e
}

func (e *exporter) registerMetric(m *MetricConfig) {
	if m.Type == "" {
		m.Type = DefaultType
	}
	name := m.Name
	desc := m.Description
	mtype := m.Type

	switch strings.ToLower(mtype) {
	case Gauge:
		e.metrics[name] = prometheus.NewGauge(prometheus.GaugeOpts{Name: name, Help: desc})
	case Counter:
		e.metrics[name] = prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: desc})
	case Summary:
		e.metrics[name] = prometheus.NewSummary(prometheus.SummaryOpts{Name: name, Help: desc})
	case Histogram:
		e.metrics[name] = prometheus.NewHistogram(prometheus.HistogramOpts{Name: name, Help: desc})
	default:
		log.Printf("ignore metric \"%v\", invalid metric type \"%v\"\n", name, mtype)
	}
	e.sources[name] = m.Source
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		ch <- m.(prometheus.Metric).Desc()
	}
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for name, metric := range e.metrics {
		source := e.sources[name]
		e.wg.Add(1)
		go e.scrape(ch, metric, source)
	}
	e.wg.Wait()
}

func (e *exporter) scrape(ch chan<- prometheus.Metric, metric interface{}, source string) {
	defer _recover()

	value := e.result(source)
	
	switch m := metric.(type) {
	case prometheus.Gauge:
		m.Set(value)
		ch <- m
	case prometheus.Counter:
		m.Add(value)
		ch <- m
	case prometheus.Summary:
		m.Observe(value)
		ch <- m
	case prometheus.Histogram:
		m.Observe(value)
		ch <- m
	}
	e.wg.Done()
}

func (e *exporter) result(source string) float64 {
	defer _recover()

	command := source
	cmd := exec.Command("/bin/bash", "-c", command)
	bytes, err := cmd.Output()
	if err != nil {
		log.Printf("execute command error, %v: %v\n", source, err)
	}
	resp := strings.TrimSpace(string(bytes))
	result, err := strconv.ParseFloat(resp, 64)
	if err != nil {
		log.Printf("parse float64 error, %v: %v\n", source, err)
	}
	return result
}

func _recover() {
	if r := recover(); r != nil {
		log.Println(string(debug.Stack()))
	}
}
