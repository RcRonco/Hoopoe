package dnsproxy

import (
	"fmt"
	"github.com/armon/go-metrics"
	prommetrics "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type TelemetryConfig struct {
	Address string `mapstructure:"Address"`
	Enabled bool
}

type TelemetryServer struct {
	config  *TelemetryConfig
}

func NewTelemetryServer(conf *TelemetryConfig) *TelemetryServer {
	telemetry := new(TelemetryServer)
	telemetry.config = conf
	telemetry.Init()
	return telemetry
}

func (s *TelemetryServer) Init() {
	metricsConfig := metrics.DefaultConfig("Hoopoe")
	metricsConfig.EnableHostnameLabel = true
	metricsConfig.EnableServiceLabel = true

	sink, err := prommetrics.NewPrometheusSink()
	handleError(err, 60)
	_, _ = metrics.NewGlobal(metricsConfig, sink)
	log.Info("Metrics: enabled.")

	http.HandleFunc("/", s.handleRoot)
	http.HandleFunc("/metrics", s.handleMetrics)
}

func (s *TelemetryServer) ListenAndServe() {
	if globalConfig.Telemetry.Enabled {
		if err := http.ListenAndServe(s.config.Address, nil); err != nil {
			handleError(err, 26)
		}
	}
}

func (s *TelemetryServer) handleRoot(resp http.ResponseWriter, req *http.Request)  {
	_, _ = fmt.Fprintf(resp, "<h1><span style=\"vertical-align: middle;\">Hoopoe</span></h1>")
}

func (s *TelemetryServer) handleMetrics(resp http.ResponseWriter, req *http.Request) {
	handlerOptions := promhttp.HandlerOpts{
		ErrorLog:      log.StandardLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}

	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, handlerOptions)
	handler.ServeHTTP(resp, req)
}