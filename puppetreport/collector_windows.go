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

func (c Collector) reportPath() string {
	if c.ReportPath != "" {
		return c.ReportPath
	}
	return "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\last_run_report.yaml"
}

func (c Collector) disabledLockPath() string {
	if c.DisabledLockPath != "" {
		return c.DisabledLockPath
	}
	return "C:\\ProgramData\\PuppetLabs\\puppet\\cache\\state\\agent_disabled.lock"
}
