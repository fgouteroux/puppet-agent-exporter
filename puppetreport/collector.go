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

package puppetreport

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	catalogVersionDesc = prometheus.NewDesc(
		"puppet_last_catalog_version_info",
		"The version of the last attempted Puppet catalog.",
		nil,
		nil,
	)

	runAtDesc = prometheus.NewDesc(
		"puppet_last_run_at_seconds",
		"Time of the last Puppet run.",
		nil,
		nil,
	)

	runDurationDesc = prometheus.NewDesc(
		"puppet_last_run_duration_seconds",
		"Duration of the last Puppet run.",
		nil,
		nil,
	)

	runSuccessDesc = prometheus.NewDesc(
		"puppet_last_run_success",
		"1 if the last Puppet run was successful.",
		nil,
		nil,
	)
	runResourcesDesc = prometheus.NewDesc(
		"puppet_last_run_report_resources",
		"Resources state of the last Puppet run",
		[]string{"type"},
		nil,
	)
	runReportTimeDurationDesc = prometheus.NewDesc(
		"puppet_last_run_report_time_duration_seconds",
		"Resources duration of the last Puppet run.",
		[]string{"type"},
		nil,
	)
	disabledLockDesc = prometheus.NewDesc(
		"puppet_disabled_lock_info",
		"Puppet state of agent disabled lock.",
		[]string{"disabled_message"},
		nil,
	)
)

type Collector struct {
	Logger           log.Logger
	ReportPath       string
	DisabledLockPath string
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- catalogVersionDesc
	ch <- runAtDesc
	ch <- runDurationDesc
	ch <- runSuccessDesc
	ch <- runResourcesDesc
	ch <- runReportTimeDurationDesc
	ch <- disabledLockDesc
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	var result interpretedReport
	if report, err := load(c.reportPath()); err != nil {
		level.Error(c.Logger).Log("msg", "Failed to read puppet run report file", "err", err)
	} else {
		result = report.interpret()
	}
	result.collect(ch)

	disabledLock, err := processDisabledLock(c.disabledLockPath())
	if err != nil {
		level.Error(c.Logger).Log("msg", "Failed to read puppet agent disabled lock file", "err", err)
	} else {
		var disabledLockMetricValue float64
		if disabledLock.Disabled {
			disabledLockMetricValue = 1
		}
		ch <- prometheus.MustNewConstMetric(disabledLockDesc, prometheus.GaugeValue, disabledLockMetricValue, []string{disabledLock.DisabledMessage}...)
	}
}

type Logger interface {
	Errorw(msg string, keysAndValues ...interface{})
}

type interpretedReport struct {
	RunAt                 float64
	RunDuration           float64
	CatalogVersion        float64
	RunSuccess            float64
	RunReportResources    map[string]float64
	RunReportTimeDuration map[string]float64
}

func (r interpretedReport) collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(catalogVersionDesc, prometheus.GaugeValue, r.CatalogVersion)
	ch <- prometheus.MustNewConstMetric(runAtDesc, prometheus.GaugeValue, r.RunAt)
	ch <- prometheus.MustNewConstMetric(runDurationDesc, prometheus.GaugeValue, r.RunDuration)
	ch <- prometheus.MustNewConstMetric(runSuccessDesc, prometheus.GaugeValue, r.RunSuccess)

	for resource, value := range r.RunReportResources {
		ch <- prometheus.MustNewConstMetric(runResourcesDesc, prometheus.GaugeValue, value, []string{resource}...)
	}

	for key, value := range r.RunReportTimeDuration {
		ch <- prometheus.MustNewConstMetric(runReportTimeDurationDesc, prometheus.GaugeValue, value, []string{key}...)
	}

}
