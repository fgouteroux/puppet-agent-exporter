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
// +build windows

package main

import (
	"log"

	"github.com/fgouteroux/puppet-agent-exporter/pkg/exporter"
	win "github.com/fgouteroux/puppet-agent-exporter/pkg/windows"
	"golang.org/x/sys/windows/svc"
)

func main() {
	e := exporter.InitExporter()

	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan bool)
	if !isInteractive {
		go func() {
			err = svc.Run("Puppet Agent Exporter", win.NewWindowsExporterService(stopCh))
			if err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
		}()
	}

	go func() {
		e.Serve()
	}()

	for {
		if <-stopCh {
			log.Printf("Shutting down %s", "Puppet Agent Exporter")
			break
		}
	}

}