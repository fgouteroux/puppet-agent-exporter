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
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"go.yaml.in/yaml/v2"
)

type runReport struct {
	ConfigurationVersion catalogVersion              `yaml:"configuration_version"`
	Time                 time.Time                   `yaml:"time"`
	TransactionCompleted bool                        `yaml:"transaction_completed"`
	ReportFormat         int                         `yaml:"report_format"`
	ResourceStatuses     map[string]resourceStatus   `yaml:"resource_statuses"`
	Metrics              map[string]puppetUtilMetric `yaml:"metrics"`
	Logs                 []puppetUtilLog             `yaml:"logs"`
}

func (r runReport) interpret() interpretedReport {
	resourcesMetrics := r.resourcesMetrics()
	return interpretedReport{
		RunAt:                 asUnixSeconds(r.Time),
		RunDuration:           r.totalDuration(),
		CatalogVersion:        float64(r.ConfigurationVersion),
		RunReportResources:    resourcesMetrics,
		RunReportEvents:       r.eventsMetrics(),
		RunReportChanges:      r.changesMetrics(),
		RunReportTimeDuration: r.reportTimeDurationMetrics(),
		RunSuccess:            r.isSuccess(resourcesMetrics),
	}
}

// catalogVersion holds Puppet's configuration_version. By default Puppet writes
// the catalog compile timestamp there, but the config_version setting may point
// at a script returning an arbitrary string. Such a report must still yield all
// the other run metrics, so a non-numeric version decodes to NaN rather than
// failing the whole document.
type catalogVersion float64

func (v *catalogVersion) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	switch value := raw.(type) {
	case nil:
		*v = catalogVersion(math.NaN())
	case int:
		*v = catalogVersion(value)
	case int64:
		*v = catalogVersion(value)
	case float64:
		*v = catalogVersion(value)
	case string:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			*v = catalogVersion(math.NaN())
			return nil
		}
		*v = catalogVersion(parsed)
	default:
		return fmt.Errorf("unsupported configuration_version type %T", raw)
	}
	return nil
}

func asUnixSeconds(t time.Time) float64 {
	return float64(t.Unix()) + (float64(t.Nanosecond()) / 1e+9)
}

// totalDuration returns the total run duration, or NaN when the report does not
// carry one. A negative duration would otherwise be indistinguishable from a
// real measurement.
func (r runReport) totalDuration() float64 {
	total, ok := r.metricValues("time")["total"]
	if !ok {
		return math.NaN()
	}
	return total
}

func (r runReport) isSuccess(resources map[string]float64) float64 {
	if !r.TransactionCompleted {
		return 0
	}

	if resources["failed"] != 0 || resources["failed_to_restart"] != 0 {
		return 0
	}

	return 1
}

// metricValues returns the values of the named puppet metric group, or an empty
// map when the report does not carry that group.
func (r runReport) metricValues(name string) map[string]float64 {
	metrics, ok := r.Metrics[name]
	if !ok {
		return make(map[string]float64)
	}
	return metrics.Values()
}

func (r runReport) resourcesMetrics() map[string]float64 {
	return r.metricValues("resources")
}

func (r runReport) eventsMetrics() map[string]float64 {
	return r.metricValues("events")
}

func (r runReport) changesMetrics() map[string]float64 {
	return r.metricValues("changes")
}

func (r runReport) reportTimeDurationMetrics() map[string]float64 {
	result := r.metricValues("time")
	// Skip total as it is already reported in RunDuration.
	delete(result, "total")
	return result
}

type resourceStatus struct {
	Failed         bool    `yaml:"failed"`
	EvaluationTime float64 `yaml:"evaluation_time"`
}

type puppetUtilMetric struct {
	Name      string     `yaml:"name"`
	Label     string     `yaml:"label"`
	RawValues [][]string `yaml:"values"`
}

func (s puppetUtilMetric) Values() map[string]float64 {
	result := make(map[string]float64, len(s.RawValues))
	for _, item := range s.RawValues {
		if len(item) == 3 {
			value, err := strconv.ParseFloat(item[2], 64)
			if err == nil {
				result[item[0]] = value
			}
		}
	}
	return result
}

type puppetUtilLog struct {
	Time time.Time `yaml:"time"`
}

func load(path string) (runReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return runReport{}, err
	}

	decoder := yaml.NewDecoder(file)
	var report runReport
	err = decoder.Decode(&report)
	return report, errors.Join(err, file.Close())
}

// reportCache memoises the parsed report. Puppet rewrites the report once per
// run (every 30 minutes by default) while the exporter may be scraped every few
// seconds, and parsing a report of a few hundred kilobytes costs tens of
// milliseconds and several megabytes of garbage each time.
type reportCache struct {
	mu       sync.Mutex
	path     string
	modTime  time.Time
	size     int64
	report   interpretedReport
	hasValue bool
}

// get returns the interpreted report for path, reusing the previous parse while
// the file's size and modification time are unchanged.
func (c *reportCache) get(path string) (interpretedReport, error) {
	info, err := os.Stat(path)
	if err != nil {
		return interpretedReport{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hasValue && c.path == path && c.size == info.Size() && c.modTime.Equal(info.ModTime()) {
		return c.report, nil
	}

	report, err := load(path)
	if err != nil {
		// Drop the stale entry so a later scrape re-reads the file rather than
		// serving a report that no longer matches what is on disk.
		c.hasValue = false
		return interpretedReport{}, err
	}

	c.path = path
	c.size = info.Size()
	c.modTime = info.ModTime()
	c.report = report.interpret()
	c.hasValue = true

	return c.report, nil
}
