package dnsproxy

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	prommetrics "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type TelemetryConfig struct {
	Enabled bool   `mapstructure:"Enabled"`
	Address string `mapstructure:"Address"`
}

type TelemetryServer struct {
	config *TelemetryConfig
}

func NewTelemetryServer(conf *TelemetryConfig) *TelemetryServer {
	telemetry := new(TelemetryServer)
	telemetry.config = conf
	return telemetry
}

func (s *TelemetryServer) Init() {
	sink, err := prommetrics.NewPrometheusSink()
	handleError(err, 60)
	_, _ = metrics.NewGlobal(metrics.DefaultConfig("Hoopoe"), sink)
	log.Info("Statistics: enabled.")

	http.HandleFunc("/", s.handleRoot)
	http.HandleFunc("/metrics", s.handleMetrics)
}

func (s *TelemetryServer) ListenAndServe() {
	if s.config.Enabled {
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