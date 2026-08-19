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

package log

import (
	"context"
	"log/slog"
	"strings"

	"github.com/prometheus/common/promslog"
	"golang.org/x/sys/windows/svc"
	el "golang.org/x/sys/windows/svc/eventlog"
)

const ServiceName = "Puppet Agent Exporter"

// IsWindowsService returns whether the current process is running as a Windows
// Service.
func IsWindowsService() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// InitLogger returns the Windows Event Logger if running as a service under
// windows, and a regular stderr logger otherwise.
func InitLogger(cfg *promslog.Config) (*slog.Logger, error) {
	if IsWindowsService() {
		return NewWindowsEventLogger(cfg)
	}
	return promslog.New(cfg), nil
}

// NewWindowsEventLogger returns a logger writing to the Windows event log.
func NewWindowsEventLogger(cfg *promslog.Config) (*slog.Logger, error) {
	// Setup the log in windows events
	err := el.InstallAsEventCreate(ServiceName, el.Error|el.Info|el.Warning)

	// Agent should expect an error of 'already exists' if the Event Log sink has already previously been installed
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, err
	}
	il, err := el.Open(ServiceName)
	if err != nil {
		return nil, err
	}

	// The handle is deliberately never closed: it is owned by the logger, which
	// lives for as long as the process does. A finalizer here could never run,
	// because the handle stays reachable from the returned logger.

	// One handler per Windows event log level, so that the promslog formatting
	// is reused: the event log API exposes a distinct call per level while an
	// io.Writer only ever sees the formatted bytes.
	newHandler := func(write func(p []byte) error) slog.Handler {
		handlerCfg := *cfg
		handlerCfg.Writer = &winLogWriter{writer: write}
		return promslog.New(&handlerCfg).Handler()
	}

	return slog.New(&winLogHandler{
		infoHandler: newHandler(func(p []byte) error {
			return il.Info(1, string(p))
		}),
		warningHandler: newHandler(func(p []byte) error {
			return il.Warning(1, string(p))
		}),
		errorHandler: newHandler(func(p []byte) error {
			return il.Error(1, string(p))
		}),
	}), nil
}

// winLogHandler dispatches records to the handler matching their level.
type winLogHandler struct {
	infoHandler    slog.Handler
	warningHandler slog.Handler
	errorHandler   slog.Handler
}

func (w *winLogHandler) handlerFor(level slog.Level) slog.Handler {
	switch {
	case level >= slog.LevelError:
		return w.errorHandler
	case level >= slog.LevelWarn:
		return w.warningHandler
	default:
		return w.infoHandler
	}
}

func (w *winLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return w.handlerFor(level).Enabled(ctx, level)
}

func (w *winLogHandler) Handle(ctx context.Context, r slog.Record) error {
	return w.handlerFor(r.Level).Handle(ctx, r)
}

func (w *winLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &winLogHandler{
		infoHandler:    w.infoHandler.WithAttrs(attrs),
		warningHandler: w.warningHandler.WithAttrs(attrs),
		errorHandler:   w.errorHandler.WithAttrs(attrs),
	}
}

func (w *winLogHandler) WithGroup(name string) slog.Handler {
	return &winLogHandler{
		infoHandler:    w.infoHandler.WithGroup(name),
		warningHandler: w.warningHandler.WithGroup(name),
		errorHandler:   w.errorHandler.WithGroup(name),
	}
}

// winLogWriter turns the formatted log line into an event log write.
type winLogWriter struct {
	writer func(p []byte) error
}

func (i *winLogWriter) Write(p []byte) (n int, err error) {
	if err := i.writer(p); err != nil {
		return 0, err
	}
	return len(p), nil
}
