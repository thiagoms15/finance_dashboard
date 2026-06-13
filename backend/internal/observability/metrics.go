package observability

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerMetricsOnce sync.Once

	apiLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "finance_api_request_duration_seconds",
			Help:    "Duration of HTTP API requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	failedLogins = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finance_auth_failed_logins_total",
			Help: "Number of failed login attempts by reason.",
		},
		[]string{"reason"},
	)
	marketRefreshTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "finance_market_data_refresh_total",
			Help: "Number of market data refresh runs by result.",
		},
		[]string{"result"},
	)
	marketRefreshDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "finance_market_data_refresh_duration_seconds",
			Help:    "Duration of market data refresh runs.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)
	marketRefreshAssets = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "finance_market_data_refresh_assets_per_run",
			Help:    "Number of asset prices updated per market data refresh run.",
			Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
	)
	marketRefreshRates = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "finance_market_data_refresh_rates_per_run",
			Help:    "Number of exchange rates updated per market data refresh run.",
			Buckets: []float64{0, 1, 2, 5, 10, 25},
		},
	)
	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "finance_db_query_duration_seconds",
			Help:    "Duration of database operations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "result"},
	)
)

func init() {
	RegisterMetrics()
}

func RegisterMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(
			apiLatency,
			failedLogins,
			marketRefreshTotal,
			marketRefreshDuration,
			marketRefreshAssets,
			marketRefreshRates,
			dbQueryDuration,
		)
	})
}

func NewLogger(service string) *slog.Logger {
	RegisterMetrics()
	return slog.New(slog.NewJSONHandler(
		gin.DefaultWriter,
		&slog.HandlerOptions{},
	)).With("service", service)
}

func MetricsHandler() http.Handler {
	RegisterMetrics()
	return promhttp.Handler()
}

func HTTPMiddleware(baseLogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start)
		apiLatency.WithLabelValues(c.Request.Method, route, status).Observe(duration.Seconds())

		RequestLogger(baseLogger, c).Info(
			"http_request_completed",
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

func RequestLogger(baseLogger *slog.Logger, c *gin.Context) *slog.Logger {
	logger := baseLogger
	if requestID, ok := c.Get("requestID"); ok {
		if id, ok := requestID.(string); ok && id != "" {
			logger = logger.With("request_id", id)
		}
	}
	if userID, ok := c.Get("userID"); ok {
		switch id := userID.(type) {
		case uuid.UUID:
			if id != uuid.Nil {
				logger = logger.With("user_id", id.String())
			}
		case string:
			if id != "" {
				logger = logger.With("user_id", id)
			}
		}
	}
	return logger
}

func RecordFailedLogin(reason string) {
	RegisterMetrics()
	if reason == "" {
		reason = "unknown"
	}
	failedLogins.WithLabelValues(reason).Inc()
}

func ObserveDBQuery(operation string, startedAt time.Time, err error) {
	RegisterMetrics()
	result := "success"
	if err != nil {
		result = "error"
	}
	dbQueryDuration.WithLabelValues(operation, result).Observe(time.Since(startedAt).Seconds())
}

func ObserveMarketRefresh(result string, duration time.Duration, assets, rates int) {
	RegisterMetrics()
	if result == "" {
		result = "unknown"
	}
	marketRefreshTotal.WithLabelValues(result).Inc()
	marketRefreshDuration.WithLabelValues(result).Observe(duration.Seconds())
	marketRefreshAssets.Observe(float64(assets))
	marketRefreshRates.Observe(float64(rates))
}
