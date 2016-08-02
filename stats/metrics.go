// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package stats

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/uber-go/zap"
)

//Metrics for the application
type Metrics struct {
	LogToStdOut        bool
	LogstashURL        string
	LogstashConnection net.Conn
	Logger             zap.Logger
	Registry           metrics.Registry
	Counters           map[string]metrics.Counter
	Histograms         map[string]metrics.Histogram
	Timers             map[string]metrics.Timer
	Quit               chan bool
}

//New Metrics endpoint
func New(logstashURL string, stdOutput bool) (*Metrics, error) {
	m := &Metrics{LogstashURL: logstashURL, LogToStdOut: stdOutput}
	m.Registry = metrics.NewRegistry()
	m.Counters = map[string]metrics.Counter{}
	m.Histograms = map[string]metrics.Histogram{}
	m.Timers = map[string]metrics.Timer{}
	m.Quit = make(chan bool)

	return m, nil
}

//IncrCounter increments a counter with the registry
func (m *Metrics) IncrCounter(name string, value int64) {
	var c metrics.Counter
	var ok bool
	if c, ok = m.Counters[name]; !ok {
		m.Counters[name] = metrics.NewCounter()
		m.Registry.Register(name, m.Counters[name])
		c = m.Counters[name]
	}
	c.Inc(value)
}

//UpdateHistogram adds a value to a histogram
func (m *Metrics) UpdateHistogram(name string, value int64) {
	var h metrics.Histogram
	var ok bool
	if h, ok = m.Histograms[name]; !ok {
		s := metrics.NewExpDecaySample(1028, 0.015)
		m.Histograms[name] = metrics.NewHistogram(s)
		m.Registry.Register(name, m.Histograms[name])
		h = m.Histograms[name]
	}
	h.Update(value)
}

//RecordTime for a given function execution
func (m *Metrics) RecordTime(name string, f func()) {
	var t metrics.Timer
	var ok bool

	if t, ok = m.Timers[name]; !ok {
		m.Timers[name] = metrics.NewTimer()
		m.Registry.Register(name, m.Timers[name])
		t = m.Timers[name]
	}
	t.Time(f)
}

//Start monitoring metrics in a goroutine
func (m *Metrics) Start() {
	if m.LogToStdOut {
		go metrics.Log(m.Registry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}

	if m.LogstashURL != "" {
		go func() {
			for {
				select {
				case <-m.Quit:
					return
				default:
					m.DumpToLogstash()
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}
}

func getCounterDetails(name string, counter metrics.Counter, ts int64) string {
	data, _ := json.Marshal(map[string]interface{}{
		"application": "khan",
		"metric":      "count",
		"name":        name,
		"timestamp":   ts,
		"value":       counter.Count(),
	})
	return string(data)
}

func getGaugeDetails(name string, gauge metrics.Gauge, ts int64) string {
	data, _ := json.Marshal(map[string]interface{}{
		"application": "khan",
		"metric":      "gauge",
		"name":        name,
		"timestamp":   ts,
		"value":       gauge.Value(),
	})
	return string(data)
}

func getFGaugeDetails(name string, gauge metrics.GaugeFloat64, ts int64) string {
	data, _ := json.Marshal(map[string]interface{}{
		"application": "khan",
		"metric":      "gauge",
		"name":        name,
		"timestamp":   ts,
		"value":       gauge.Value(),
	})
	return string(data)
}
func getHistogramDetails(name string, metric metrics.Histogram, ts int64) string {
	h := metric.Snapshot()
	ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})

	data, _ := json.Marshal(map[string]interface{}{
		"application":   "khan",
		"metric":        "gauge",
		"name":          name,
		"timestamp":     ts,
		"count":         h.Count(),
		"min":           h.Min(),
		"max":           h.Max(),
		"mean":          h.Mean(),
		"stdDev":        h.StdDev(),
		"percentile50":  ps[0],
		"percentile75":  ps[1],
		"percentile95":  ps[2],
		"percentile99":  ps[3],
		"percentile999": ps[4],
	})

	return string(data)
}

func getTimerDetails(name string, metric metrics.Timer, ts int64) string {
	t := metric.Snapshot()
	ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})

	data, _ := json.Marshal(map[string]interface{}{
		"application":   "khan",
		"metric":        "gauge",
		"name":          name,
		"timestamp":     ts,
		"count":         t.Count(),
		"min":           t.Min(),
		"max":           t.Max(),
		"mean":          t.Mean(),
		"stdDev":        t.StdDev(),
		"percentile50":  ps[0],
		"percentile75":  ps[1],
		"percentile95":  ps[2],
		"percentile99":  ps[3],
		"percentile999": ps[4],
		"oneMinute":     t.Rate1(),
		"fiveMinute":    t.Rate5(),
		"fifteenMinute": t.Rate15(),
		"meanRate":      t.RateMean(),
	})

	return string(data)
}

//DumpToLogstash all metrics
func (m *Metrics) DumpToLogstash() {
	now := time.Now().Unix()

	conn, err := net.Dial("udp", m.LogstashURL)
	if err != nil {
		m.Logger.Error("Could not connect to logstash.", zap.String("LogstashURL", m.LogstashURL), zap.Error(err))
		return
	}
	defer conn.Close()
	m.LogstashConnection = conn

	defer conn.Close()
	w := bufio.NewWriter(conn)
	m.Registry.Each(func(name string, i interface{}) {
		var repr string
		switch metric := i.(type) {
		case metrics.Counter:
			repr = getCounterDetails(name, metric, now)
		case metrics.Gauge:
			repr = getGaugeDetails(name, metric, now)
		case metrics.GaugeFloat64:
			repr = getFGaugeDetails(name, metric, now)
		case metrics.Histogram:
			repr = getHistogramDetails(name, metric, now)
		//case metrics.Meter:
		//m := metric.Snapshot()
		//repr = "meter.%s.count %d %d\n", name, now, m.Count())
		//repr = "meter.%s.one-minute %d %.2f\n", name, now, m.Rate1())
		//repr = "meter.%s.five-minute %d %.2f\n", name, now, m.Rate5())
		//repr = "meter.%s.fifteen-minute %d %.2f\n", name, now, m.Rate15())
		//repr = "meter.%s.mean %d %.2f\n", name, now, m.RateMean())
		case metrics.Timer:
			repr = getTimerDetails(name, metric, now)
		}
		fmt.Fprintf(w, repr)
		w.Flush()
	})
}

//Stop monitoring metrics
func (m *Metrics) Stop() {
	if m.LogstashConnection != nil {
		m.Quit <- true
		close(m.Quit)
		m.LogstashConnection.Close()
		m.LogstashConnection = nil
	}
}
