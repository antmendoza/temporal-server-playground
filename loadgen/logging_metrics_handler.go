package loadgen

import (
	"log"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
)

// loggingMetricsHandler wraps a client.MetricsHandler and prints every
// record/update call for metrics whose name contains filter.
// Pass an empty filter to log all metrics.
type loggingMetricsHandler struct {
	inner  client.MetricsHandler
	tags   map[string]string
	filter string
}

func NewLoggingMetricsHandler(inner client.MetricsHandler, filter string) client.MetricsHandler {
	return &loggingMetricsHandler{inner: inner, filter: filter}
}

func (h *loggingMetricsHandler) matches(name string) bool {
	return h.filter == "" || strings.Contains(name, h.filter)
}

func (h *loggingMetricsHandler) WithTags(tags map[string]string) client.MetricsHandler {
	merged := make(map[string]string, len(h.tags)+len(tags))
	for k, v := range h.tags {
		merged[k] = v
	}
	for k, v := range tags {
		merged[k] = v
	}
	return &loggingMetricsHandler{inner: h.inner.WithTags(tags), tags: merged, filter: h.filter}
}

type loggingCounter struct {
	inner client.MetricsCounter
	name  string
	tags  map[string]string
}

func (c *loggingCounter) Inc(v int64) {
	log.Printf("[metrics] counter  %-55s  value=%d  tags=%v", c.name, v, c.tags)
	c.inner.Inc(v)
}

func (h *loggingMetricsHandler) Counter(name string) client.MetricsCounter {
	inner := h.inner.Counter(name)
	if !h.matches(name) {
		return inner
	}
	return &loggingCounter{inner: inner, name: name, tags: h.tags}
}

type loggingGauge struct {
	inner client.MetricsGauge
	name  string
	tags  map[string]string
}

func (g *loggingGauge) Update(v float64) {
	log.Printf("[metrics] gauge    %-55s  value=%g  tags=%v", g.name, v, g.tags)
	g.inner.Update(v)
}

func (h *loggingMetricsHandler) Gauge(name string) client.MetricsGauge {
	inner := h.inner.Gauge(name)
	if !h.matches(name) {
		return inner
	}
	return &loggingGauge{inner: inner, name: name, tags: h.tags}
}

type loggingTimer struct {
	inner client.MetricsTimer
	name  string
	tags  map[string]string
}

func (t *loggingTimer) Record(v time.Duration) {
	log.Printf("[metrics] timer    %-55s  value=%s  tags=%v", t.name, v, t.tags)
	t.inner.Record(v)
}

func (h *loggingMetricsHandler) Timer(name string) client.MetricsTimer {
	inner := h.inner.Timer(name)
	if !h.matches(name) {
		return inner
	}
	return &loggingTimer{inner: inner, name: name, tags: h.tags}
}
