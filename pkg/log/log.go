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

//go:build !windows

package log

import (
	"log/slog"

	"github.com/prometheus/common/promslog"
)

// InitLogger returns a logger writing to standard error.
func InitLogger(cfg *promslog.Config) (*slog.Logger, error) {
	return promslog.New(cfg), nil
}
