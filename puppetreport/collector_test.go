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
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

func writeReport(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "last_run_report.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Puppet's config_version setting may point at a script returning an arbitrary
// string. That must not cost us every other metric in the report.
func TestNonNumericCatalogVersion(t *testing.T) {
	path := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: my-release-2024.1
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
metrics:
  resources:
    name: resources
    values:
    - - total
      - Total
      - 12
    - - failed
      - Failed
      - 0
  time:
    name: time
    values:
    - - total
      - Total
      - 17.5
`)

	report, err := load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	ir := report.interpret()
	if !math.IsNaN(ir.CatalogVersion) {
		t.Errorf("CatalogVersion = %v, want NaN", ir.CatalogVersion)
	}
	if ir.RunSuccess != 1 {
		t.Errorf("RunSuccess = %v, want 1", ir.RunSuccess)
	}
	if ir.RunDuration != 17.5 {
		t.Errorf("RunDuration = %v, want 17.5", ir.RunDuration)
	}
	if got := ir.RunReportResources["total"]; got != 12 {
		t.Errorf("resources[total] = %v, want 12", got)
	}
}

func TestNumericCatalogVersionStillParses(t *testing.T) {
	report, err := load("last_run_report.yaml")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := report.interpret().CatalogVersion; got != 1618957129 {
		t.Errorf("CatalogVersion = %v, want 1618957129", got)
	}
}

// A report without a time metric group has no duration to report; NaN says so,
// where the previous -1 was indistinguishable from a measurement.
func TestMissingDurationIsNaN(t *testing.T) {
	path := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 1618957129
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`)

	report, err := load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := report.interpret().RunDuration; !math.IsNaN(got) {
		t.Errorf("RunDuration = %v, want NaN", got)
	}
}

func TestFailedRunIsNotSuccess(t *testing.T) {
	path := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 1618957129
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
metrics:
  resources:
    name: resources
    values:
    - - failed
      - Failed
      - 3
`)

	report, err := load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := report.interpret().RunSuccess; got != 0 {
		t.Errorf("RunSuccess = %v, want 0", got)
	}
}

func TestCollectMissingFile(t *testing.T) {
	c := &Collector{Logger: promslog.NewNopLogger(), ReportPath: filepath.Join(t.TempDir(), "absent.yaml")}

	expected := `
# HELP puppet_last_run_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_last_run_scrape_error gauge
puppet_last_run_scrape_error 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "puppet_last_run_scrape_error"); err != nil {
		t.Fatal(err)
	}
}

func TestDescribeCoversCollect(t *testing.T) {
	c := &Collector{Logger: promslog.NewNopLogger(), ReportPath: "last_run_report.yaml"}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("pedantic gather: %v", err)
	}
}

func TestCacheReusesParseUntilFileChanges(t *testing.T) {
	path := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 1
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`)

	var cache reportCache
	if first, err := cache.get(path); err != nil {
		t.Fatal(err)
	} else if first.CatalogVersion != 1 {
		t.Fatalf("CatalogVersion = %v, want 1", first.CatalogVersion)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	// Same size, same mtime: the cached parse is reused, so the new contents are
	// deliberately not observed. This is the trade-off the cache makes, and it
	// is what proves the file was not re-read.
	if err := os.WriteFile(path, []byte(`--- !ruby/object:Puppet::Transaction::Report
configuration_version: 9
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if cached, err := cache.get(path); err != nil {
		t.Fatal(err)
	} else if cached.CatalogVersion != 1 {
		t.Errorf("CatalogVersion = %v, want the cached 1", cached.CatalogVersion)
	}

	// A newer mtime, as a real Puppet run produces, invalidates the entry.
	if err := os.Chtimes(path, info.ModTime(), info.ModTime().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if refreshed, err := cache.get(path); err != nil {
		t.Fatal(err)
	} else if refreshed.CatalogVersion != 9 {
		t.Errorf("CatalogVersion = %v after the file changed, want 9", refreshed.CatalogVersion)
	}
}

// A different path must not be served from the previous file's entry.
func TestCacheKeyedOnPath(t *testing.T) {
	first := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 1
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`)
	second := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 2
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`)
	if err := os.Chtimes(second, time.Time{}, mustModTime(t, first)); err != nil {
		t.Fatal(err)
	}

	var cache reportCache
	if _, err := cache.get(first); err != nil {
		t.Fatal(err)
	}
	got, err := cache.get(second)
	if err != nil {
		t.Fatal(err)
	}
	if got.CatalogVersion != 2 {
		t.Errorf("CatalogVersion = %v, want 2 from the second path", got.CatalogVersion)
	}
}

func mustModTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func TestCacheDoesNotServeStaleAfterError(t *testing.T) {
	path := writeReport(t, `--- !ruby/object:Puppet::Transaction::Report
configuration_version: 1
time: '2021-04-20T22:18:45.590110290+00:00'
transaction_completed: true
`)

	var cache reportCache
	if _, err := cache.get(path); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.get(path); err == nil {
		t.Fatal("expected an error once the report is gone")
	}
}
