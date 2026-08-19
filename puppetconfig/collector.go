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
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/ini.v1"
)

var (
	configDesc = prometheus.NewDesc(
		"puppet_config",
		"Puppet configuration.",
		[]string{"server", "environment"},
		nil,
	)
	scrapeErrorDesc = prometheus.NewDesc(
		"puppet_config_scrape_error",
		"1 if there was an error opening or reading a file, 0 otherwise",
		nil,
		nil,
	)
)

// sections are searched in the order Puppet itself resolves agent settings:
// the agent-specific section wins over the global one.
var sections = []string{"agent", "main"}

type Collector struct {
	Logger     *slog.Logger
	ConfigPath string
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- configDesc
	ch <- scrapeErrorDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var errVal float64
	config, err := ini.Load(c.configPath())
	if err != nil {
		c.Logger.Error("Failed to open puppet config file", "err", err)
		errVal = 1.0
	} else {
		server := setting(config, "server")
		environment := setting(config, "environment")
		ch <- prometheus.MustNewConstMetric(configDesc, prometheus.GaugeValue, 1, server, environment)
	}

	ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, errVal)
}

// setting returns the value of an agent setting, honouring the section
// precedence Puppet applies when the agent reads its own configuration.
func setting(config *ini.File, key string) string {
	for _, section := range sections {
		if value := config.Section(section).Key(key).String(); value != "" {
			return value
		}
	}
	return ""
}
