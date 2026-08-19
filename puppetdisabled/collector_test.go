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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promslog"
)

const scrapeErrorHeader = `
# HELP puppet_disabled_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE puppet_disabled_scrape_error gauge
`

func lockFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agent_disabled.lock")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCollect(t *testing.T) {
	for _, tc := range []struct {
		name     string
		path     func(t *testing.T) string
		expected string
	}{
		{
			name: "not disabled",
			path: func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent.lock") },
			expected: `
# HELP puppet_disabled_lock_info Puppet state of agent disabled lock.
# TYPE puppet_disabled_lock_info gauge
puppet_disabled_lock_info{disabled_message=""} 0
` + scrapeErrorHeader + `puppet_disabled_scrape_error 0
`,
		},
		{
			name: "disabled with message",
			path: func(t *testing.T) string { return lockFile(t, `{"disabled_message":"maintenance window"}`) },
			expected: `
# HELP puppet_disabled_lock_info Puppet state of agent disabled lock.
# TYPE puppet_disabled_lock_info gauge
puppet_disabled_lock_info{disabled_message="maintenance window"} 1
` + scrapeErrorHeader + `puppet_disabled_scrape_error 0
`,
		},
		{
			// The lock file existing is what disables the agent, so an
			// unparseable body must still report the agent as disabled.
			name: "disabled with unparseable body",
			path: func(t *testing.T) string { return lockFile(t, "") },
			expected: `
# HELP puppet_disabled_lock_info Puppet state of agent disabled lock.
# TYPE puppet_disabled_lock_info gauge
puppet_disabled_lock_info{disabled_message=""} 1
` + scrapeErrorHeader + `puppet_disabled_scrape_error 1
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Collector{Logger: promslog.NewNopLogger(), LockPath: tc.path(t)}
			if err := testutil.CollectAndCompare(c, strings.NewReader(tc.expected)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A lock file that cannot be read at all leaves the agent state unknown, which
// must not be reported as either enabled or disabled.
func TestCollectUnreadableLeavesStateUnknown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "agent_disabled.lock")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(dir, "sub"), 0o700) })

	if _, err := os.ReadFile(path); err == nil || os.IsNotExist(err) {
		t.Skip("cannot produce an unreadable lock file (running as root?)")
	}

	c := &Collector{Logger: promslog.NewNopLogger(), LockPath: path}
	expected := scrapeErrorHeader + "puppet_disabled_scrape_error 1\n"
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected)); err != nil {
		t.Fatal(err)
	}
}

func TestTruncate(t *testing.T) {
	long := strings.Repeat("x", maxDisabledMessageLen+50)
	if got := truncate(long); len(got) != maxDisabledMessageLen {
		t.Fatalf("truncate(%d chars) = %d chars, want %d", len(long), len(got), maxDisabledMessageLen)
	}
	if got := truncate("short"); got != "short" {
		t.Fatalf("truncate(%q) = %q", "short", got)
	}
}

func TestDescribeCoversCollect(t *testing.T) {
	c := &Collector{Logger: promslog.NewNopLogger(), LockPath: lockFile(t, `{"disabled_message":"why"}`)}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("pedantic gather: %v", err)
	}
}
