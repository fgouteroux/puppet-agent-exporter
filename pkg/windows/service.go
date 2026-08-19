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

//go:build windows

package windows

import (
	"log/slog"

	"golang.org/x/sys/windows/svc"
)

// ExporterService handles the Windows service control requests and signals the
// exporter to stop.
type ExporterService struct {
	stopCh chan<- bool
	logger *slog.Logger
}

// NewExporterService returns a new ExporterService.
func NewExporterService(ch chan<- bool, logger *slog.Logger) *ExporterService {
	return &ExporterService{stopCh: ch, logger: logger}
}

// Execute runs the service control loop until a stop or shutdown is requested.
func (s *ExporterService) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s.stopCh <- true
			break loop
		default:
			s.logger.Warn("Unexpected service control request", "cmd", c.Cmd)
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}
