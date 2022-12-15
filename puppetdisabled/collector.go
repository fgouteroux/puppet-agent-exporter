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

package puppetdisabled

import (
	"encoding/json"
	"errors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"os"
)

var disabledLockDesc = prometheus.NewDesc(
	"puppet_disabled_lock_info",
	"Puppet state of agent disabled lock.",
	[]string{"disabled_message"},
	nil,
)

type Collector struct {
	Logger   log.Logger
	LockPath string
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- disabledLockDesc
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	var errVal float64
	disabledLock, err := processDisabledLock(c.lockPath())
	if err != nil {
		level.Error(c.Logger).Log("msg", "Failed to read puppet agent disabled lock file", "err", err)
		errVal = 1.0
	} else {
		var disabledLockMetricValue float64
		if disabledLock.Disabled {
			disabledLockMetricValue = 1
		}
		ch <- prometheus.MustNewConstMetric(disabledLockDesc, prometheus.GaugeValue, disabledLockMetricValue, []string{disabledLock.DisabledMessage}...)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"puppet_disabled_scrape_error",
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, errVal,
	)
}

type agentDisabledLock struct {
	Disabled        bool
	DisabledMessage string `json:"disabled_message"`
}

func processDisabledLock(path string) (agentDisabledLock, error) {
	disabledLockContent, err := os.ReadFile(path)

	var disabledLock agentDisabledLock
	disabledLock.Disabled = true

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			disabledLock.Disabled = false
		} else {
			return disabledLock, err
		}
	}

	if disabledLockContent != nil {
		if err := json.Unmarshal(disabledLockContent, &disabledLock); err != nil {
			return disabledLock, err
		}
	}

	return disabledLock, nil

}
