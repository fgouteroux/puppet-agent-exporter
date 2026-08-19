# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO         ?= go
GORELEASER ?= goreleaser

GOLANGCI_LINT_VERSION ?= v2.12.2
GOLANGCI_LINT         ?= golangci-lint

# Whenever this is updated, .github/workflows/go.yml should also be updated.
PROMETHEUS_VERSION ?= 3.14.0
PROMTOOL           ?= /tmp/prometheus-$(PROMETHEUS_VERSION).linux-amd64/promtool

.PHONY: all
all: style vet lint test checkmetrics build

.PHONY: style
style:
	@echo ">> checking code style"
	@fmtRes=$$($(GO) fmt ./...); \
	if [ -n "$${fmtRes}" ]; then \
		echo "gofmt checking failed:"; echo "$${fmtRes}"; echo; \
		exit 1; \
	fi

.PHONY: vet
vet:
	@echo ">> vetting code"
	$(GO) vet ./...
	GOOS=windows $(GO) vet ./...

.PHONY: lint
lint:
	@echo ">> running golangci-lint"
	$(GOLANGCI_LINT) run ./...
	GOOS=windows $(GOLANGCI_LINT) run ./...

.PHONY: test
test:
	@echo ">> running tests"
	$(GO) test -race ./...

.PHONY: checkmetrics
checkmetrics: $(PROMTOOL)
	@echo ">> checking metrics for correctness"
	for file in test/*.metrics; do $(PROMTOOL) check metrics < $$file || exit 1; done

$(PROMTOOL):
	curl -sL -o - https://github.com/prometheus/prometheus/releases/download/v$(PROMETHEUS_VERSION)/prometheus-$(PROMETHEUS_VERSION).linux-amd64.tar.gz \
		| tar -C /tmp -xzf - prometheus-$(PROMETHEUS_VERSION).linux-amd64/promtool

# Cross-compiles every release target and builds the packages, without
# publishing anything. This is what CI runs on pull requests.
.PHONY: build
build:
	@echo ">> building release artefacts (snapshot)"
	$(GORELEASER) release --snapshot --clean --skip=publish

.PHONY: tidy
tidy:
	$(GO) mod tidy
	@git diff --exit-code -- go.sum go.mod

.PHONY: print-golangci-lint-version
print-golangci-lint-version:
	@echo $(GOLANGCI_LINT_VERSION)

.PHONY: clean
clean:
	rm -rf dist
