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

package puppetconfig

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/ini.v1"
)

var configDesc = prometheus.NewDesc(
	"puppet_config",
	"Puppet configuration.",
	[]string{"server", "environment"},
	nil,
)

type Collector struct {
	Logger     log.Logger
	ConfigPath string
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- configDesc
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	var errVal float64
	config, err := ini.Load(c.configPath())
	if err != nil {
		level.Error(c.Logger).Log("msg", "Failed to open puppet config file", "err", err)
		errVal = 1.0
	} else {
		server := config.Section("main").Key("server").String()
		environment := config.Section("main").Key("environment").String()
		ch <- prometheus.MustNewConstMetric(configDesc, prometheus.GaugeValue, 1, server, environment)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"puppet_config_scrape_error",
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, errVal,
	)
}
