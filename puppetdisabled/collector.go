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
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	disabledLockDesc = prometheus.NewDesc(
		"puppet_disabled_lock_info",
		"Puppet state of agent disabled lock.",
		[]string{"disabled_message"},
		nil,
	)
	scrapeErrorDesc = prometheus.NewDesc(
		"puppet_disabled_scrape_error",
		"1 if there was an error opening or reading a file, 0 otherwise",
		nil,
		nil,
	)
)

// maxDisabledMessageLen bounds the free-form operator text that ends up as a
// label value, so a pathological message cannot blow up series size.
const maxDisabledMessageLen = 128

type Collector struct {
	Logger   *slog.Logger
	LockPath string
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- disabledLockDesc
	ch <- scrapeErrorDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var errVal float64
	disabledLock, err := processDisabledLock(c.lockPath())
	if err != nil {
		c.Logger.Error("Failed to read puppet agent disabled lock file", "err", err)
		errVal = 1.0
	}

	// The presence of the lock file is what disables the agent, so the state is
	// known even when the message inside it could not be parsed. Report it
	// whenever it is known, and let puppet_disabled_scrape_error flag the parse
	// failure separately.
	if disabledLock.stateKnown {
		var disabledLockMetricValue float64
		if disabledLock.Disabled {
			disabledLockMetricValue = 1
		}
		ch <- prometheus.MustNewConstMetric(disabledLockDesc, prometheus.GaugeValue, disabledLockMetricValue, truncate(disabledLock.DisabledMessage))
	}

	ch <- prometheus.MustNewConstMetric(scrapeErrorDesc, prometheus.GaugeValue, errVal)
}

func truncate(message string) string {
	if len(message) <= maxDisabledMessageLen {
		return message
	}
	return message[:maxDisabledMessageLen]
}

type agentDisabledLock struct {
	Disabled        bool
	DisabledMessage string `json:"disabled_message"`

	// stateKnown records whether the lock file could be inspected at all. A
	// read failure leaves the agent state unknown, which must not be reported
	// as either enabled or disabled.
	stateKnown bool
}

// processDisabledLock reports whether the agent is disabled, and why. The lock
// file existing is what disables the agent; its JSON body only carries the
// message. A body that cannot be parsed therefore still means disabled, while a
// file that cannot be read at all leaves the state unknown.
func processDisabledLock(path string) (agentDisabledLock, error) {
	disabledLockContent, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return agentDisabledLock{Disabled: false, stateKnown: true}, nil
		}
		return agentDisabledLock{}, err
	}

	disabledLock := agentDisabledLock{Disabled: true, stateKnown: true}
	if err := json.Unmarshal(disabledLockContent, &disabledLock); err != nil {
		// Keep Disabled set: the file is there, only the message is unusable.
		return agentDisabledLock{Disabled: true, stateKnown: true}, err
	}

	return disabledLock, nil
}
