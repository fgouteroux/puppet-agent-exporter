// Copyright 2021 RetailNext, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	promslogflag "github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	customlog "github.com/fgouteroux/puppet-agent-exporter/pkg/log"
	"github.com/fgouteroux/puppet-agent-exporter/puppetconfig"
	"github.com/fgouteroux/puppet-agent-exporter/puppetdisabled"
	"github.com/fgouteroux/puppet-agent-exporter/puppetreport"
)

type Exporter struct {
	server    *http.Server
	Logger    *slog.Logger
	webConfig *web.FlagConfig
}

func InitExporter() (e *Exporter) {
	var (
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":9819")
	)
	promslogConfig := &promslog.Config{}
	promslogflag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("puppet-agent-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger, err := customlog.InitLogger(promslogConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init custom logger: %v\n", err)
		os.Exit(1)
	}

	prometheus.MustRegister(puppetconfig.Collector{
		Logger: logger,
	})
	prometheus.MustRegister(puppetreport.Collector{
		Logger: logger,
	})
	prometheus.MustRegister(puppetdisabled.Collector{
		Logger: logger,
	})
	prometheus.MustRegister(versioncollector.NewCollector("puppet_agent_exporter"))

	logger.Info("Starting puppet-agent-exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`<html>
			<head><title>Puppet Agent Exporter</title></head>
			<body>
			<h1>Puppet Agent Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`)); err != nil {
			logger.Error("Failed to write landing page", "err", err)
		}
	})

	return &Exporter{
		server:    &http.Server{Handler: mux},
		Logger:    logger,
		webConfig: webConfig,
	}
}

// Serve Start the http web server
func (e *Exporter) Serve() {
	if err := web.ListenAndServe(e.server, e.webConfig, e.Logger); err != nil {
		e.Logger.Error("Failed to run web server", "err", err)
		os.Exit(1)
	}
}
