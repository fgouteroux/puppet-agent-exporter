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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "puppet.conf")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCollect(t *testing.T) {
	for _, tc := range []struct {
		name     string
		config   string
		expected string
	}{
		{
			name:   "settings in main",
			config: "[main]\nserver = puppet.example.com\nenvironment = production\n",
			expected: `
# HELP puppet_config Puppet configuration.
# TYPE puppet_config gauge
puppet_config{environment="production",server="puppet.example.com"} 1
# HELP puppet_config_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_config_scrape_error gauge
puppet_config_scrape_error 0
`,
		},
		{
			// Puppet resolves agent settings from [agent] as well as [main].
			name:   "settings in agent",
			config: "[main]\nvardir = /opt/puppetlabs/puppet/cache\n\n[agent]\nserver = puppet.example.com\nenvironment = production\n",
			expected: `
# HELP puppet_config Puppet configuration.
# TYPE puppet_config gauge
puppet_config{environment="production",server="puppet.example.com"} 1
# HELP puppet_config_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_config_scrape_error gauge
puppet_config_scrape_error 0
`,
		},
		{
			name:   "agent overrides main",
			config: "[main]\nserver = old.example.com\nenvironment = staging\n\n[agent]\nserver = puppet.example.com\nenvironment = production\n",
			expected: `
# HELP puppet_config Puppet configuration.
# TYPE puppet_config gauge
puppet_config{environment="production",server="puppet.example.com"} 1
# HELP puppet_config_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_config_scrape_error gauge
puppet_config_scrape_error 0
`,
		},
		{
			name:   "no settings",
			config: "# nothing configured\n",
			expected: `
# HELP puppet_config Puppet configuration.
# TYPE puppet_config gauge
puppet_config{environment="",server=""} 1
# HELP puppet_config_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_config_scrape_error gauge
puppet_config_scrape_error 0
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Collector{Logger: promslog.NewNopLogger(), ConfigPath: writeConfig(t, tc.config)}
			if err := testutil.CollectAndCompare(c, strings.NewReader(tc.expected)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCollectMissingFile(t *testing.T) {
	c := &Collector{Logger: promslog.NewNopLogger(), ConfigPath: filepath.Join(t.TempDir(), "absent.conf")}

	expected := `
# HELP puppet_config_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_config_scrape_error gauge
puppet_config_scrape_error 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected)); err != nil {
		t.Fatal(err)
	}
}

// Every descriptor emitted by Collect must be announced by Describe, otherwise a
// pedantic registry refuses to gather.
func TestDescribeCoversCollect(t *testing.T) {
	c := &Collector{Logger: promslog.NewNopLogger(), ConfigPath: writeConfig(t, "[main]\nserver = puppet.example.com\n")}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("pedantic gather: %v", err)
	}
}
