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
		"puppet_last_run_catalog_version",
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
	runEventsDesc = prometheus.NewDesc(
		"puppet_last_run_report_events",
		"Events state of the last Puppet run",
		[]string{"type"},
		nil,
	)
	runChangesDesc = prometheus.NewDesc(
		"puppet_last_run_report_changes",
		"Changes of the last Puppet run",
		[]string{"type"},
		nil,
	)
	runReportTimeDurationDesc = prometheus.NewDesc(
		"puppet_last_run_report_time_duration_seconds",
		"Resources duration of the last Puppet run.",
		[]string{"type"},
		nil,
	)
)

type Collector struct {
	Logger     log.Logger
	ReportPath string
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- catalogVersionDesc
	ch <- runAtDesc
	ch <- runDurationDesc
	ch <- runSuccessDesc
	ch <- runResourcesDesc
	ch <- runEventsDesc
	ch <- runChangesDesc
	ch <- runReportTimeDurationDesc
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	var errVal float64
	if report, err := load(c.reportPath()); err != nil {
		level.Error(c.Logger).Log("msg", "Failed to read puppet run report file", "err", err)
		errVal = 1.0
	} else {
		result := report.interpret()
		result.collect(ch)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"puppet_last_run_scrape_error",
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, errVal,
	)
}

type interpretedReport struct {
	RunAt                 float64
	RunDuration           float64
	CatalogVersion        float64
	RunSuccess            float64
	RunReportResources    map[string]float64
	RunReportEvents       map[string]float64
	RunReportChanges      map[string]float64
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

	for event, value := range r.RunReportEvents {
		ch <- prometheus.MustNewConstMetric(runEventsDesc, prometheus.GaugeValue, value, []string{event}...)
	}

	for change, value := range r.RunReportChanges {
		ch <- prometheus.MustNewConstMetric(runChangesDesc, prometheus.GaugeValue, value, []string{change}...)
	}

	for key, value := range r.RunReportTimeDuration {
		ch <- prometheus.MustNewConstMetric(runReportTimeDurationDesc, prometheus.GaugeValue, value, []string{key}...)
	}

}
