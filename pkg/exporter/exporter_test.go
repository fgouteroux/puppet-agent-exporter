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

package exporter

import "testing"

func TestValidateTelemetryPath(t *testing.T) {
	for _, tc := range []struct {
		path    string
		wantErr bool
	}{
		{path: "/metrics"},
		{path: "/puppet/metrics"},
		// "/" and "" used to panic the process inside net/http's ServeMux.
		{path: "/", wantErr: true},
		{path: "", wantErr: true},
		{path: "metrics", wantErr: true},
	} {
		err := validateTelemetryPath(tc.path)
		if tc.wantErr && err == nil {
			t.Errorf("validateTelemetryPath(%q) = nil, want an error", tc.path)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("validateTelemetryPath(%q) = %v, want nil", tc.path, err)
		}
	}
}
