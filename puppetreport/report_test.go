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
	"reflect"
	"testing"
)

func TestLoadReport(t *testing.T) {
	report, err := load("last_run_report.yaml")
	if err != nil {
		t.Fatal(err)
	}

	ir := report.interpret()
	expected := interpretedReport{
		RunAt:          1618957125.5901103,
		RunDuration:    17.199882286,
		CatalogVersion: 1618957129,
		RunSuccess:     1,
		RunReportResources: map[string]float64{
			"total":             574,
			"skipped":           0,
			"failed":            0,
			"failed_to_restart": 0,
			"restarted":         0,
			"changed":           1,
			"out_of_sync":       1,
			"scheduled":         0,
			"corrective_change": 1,
		},
		RunReportTimeDuration: map[string]float64{
			"plugin_sync":            19.038186447694898,
			"fact_generation":        3.733549404889345,
			"convert_catalog":        1.6039954144507647,
			"config_retrieval":       7.887831624597311,
			"transaction_evaluation": 23.296303944662213,
			"catalog_application":    23.389429319649935,
		},
	}

	if !reflect.DeepEqual(ir, expected) {
		t.Fatalf("%+v != %+v", ir, expected)
	}
}
