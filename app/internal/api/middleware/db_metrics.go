package middleware

import (
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
)

type dbStatsCollector struct {
	db *sqlx.DB

	maxOpen     *prometheus.Desc
	open        *prometheus.Desc
	inUse       *prometheus.Desc
	idle        *prometheus.Desc
	waitCount   *prometheus.Desc
	waitSeconds *prometheus.Desc
}

// These were written by Claude.  They seem to work, but I need to take a closer look at them later

// RegisterDBMetrics creates and registers a collector for database connection pool stats.
func RegisterDBMetrics(db *sqlx.DB) {
	collector := &dbStatsCollector{
		db: db,
		maxOpen: prometheus.NewDesc(
			"syp_db_max_open_connections",
			"Maximum number of open connections to the database.",
			nil, nil,
		),
		open: prometheus.NewDesc(
			"syp_db_open_connections",
			"Current number of open connections to the database.",
			nil, nil,
		),
		inUse: prometheus.NewDesc(
			"syp_db_in_use_connections",
			"Number of connections currently in use.",
			nil, nil,
		),
		idle: prometheus.NewDesc(
			"syp_db_idle_connections",
			"Number of idle connections.",
			nil, nil,
		),
		waitCount: prometheus.NewDesc(
			"syp_db_wait_count_total",
			"Total number of connections waited for.",
			nil, nil,
		),
		waitSeconds: prometheus.NewDesc(
			"syp_db_wait_seconds_total",
			"Total time blocked waiting for a new connection.",
			nil, nil,
		),
	}
	prometheus.MustRegister(collector)
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpen
	ch <- c.open
	ch <- c.inUse
	ch <- c.idle
	ch <- c.waitCount
	ch <- c.waitSeconds
}

func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(c.maxOpen, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
	ch <- prometheus.MustNewConstMetric(c.open, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.inUse, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(c.idle, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.waitSeconds, prometheus.CounterValue, stats.WaitDuration.Seconds())
}
