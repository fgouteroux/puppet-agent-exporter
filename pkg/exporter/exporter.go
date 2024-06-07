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
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	customlog "github.com/fgouteroux/puppet-agent-exporter/pkg/log"
	"github.com/fgouteroux/puppet-agent-exporter/puppetconfig"
	"github.com/fgouteroux/puppet-agent-exporter/puppetdisabled"
	"github.com/fgouteroux/puppet-agent-exporter/puppetreport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Exporter struct {
	server    *http.Server
	Logger    log.Logger
	webConfig *web.FlagConfig
}

func InitExporter() (e *Exporter) {
	var (
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":9819")
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("puppet-agent-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger, err := customlog.InitLogger(promlogConfig)
	if err != nil {
		var logger log.Logger
		level.Error(logger).Log("Failed to init custom logger", err)
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
	prometheus.MustRegister(version.NewCollector("puppet_agent_exporter"))

	level.Info(logger).Log("msg", "Starting puppet-agent-exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Puppet Agent Exporter</title></head>
			<body>
			<h1>Puppet Agent Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	return &Exporter{
		server:    &http.Server{},
		Logger:    logger,
		webConfig: webConfig,
	}
}

// Serve Start the http web server
func (e *Exporter) Serve() {
	if err := web.ListenAndServe(e.server, e.webConfig, e.Logger); err != nil {
		level.Error(e.Logger).Log("err", err)
		os.Exit(1)
	}
}
