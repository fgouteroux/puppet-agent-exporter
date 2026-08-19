## Unreleased

* [CHANGE] replace promu/promci with GoReleaser for building and releasing
* [CHANGE] drop Makefile.common in favour of a self-contained Makefile
* [FEATURE] publish deb and rpm packages with a systemd unit
* [FEATURE] build linux/arm64 and darwin/arm64 artefacts

## 0.1.7 / 2026-08-19

* [CHANGE] go 1.26
* [CHANGE] update all go dependencies
* [CHANGE] replace go-kit/log with the standard library log/slog
* [FEATURE] add --puppet.config-path, --puppet.lock-path and --puppet.report-path
* [ENHANCEMENT] cache the parsed run report until the file changes
* [ENHANCEMENT] declare the scrape_error metrics in Describe
* [ENHANCEMENT] limit the length of the disabled_message label
* [ENHANCEMENT] set a read header timeout on the web server
* [FIX] keep reporting run metrics when configuration_version is not numeric
* [FIX] read server and environment from the puppet.conf [agent] section
* [FIX] report the disabled lock state when its message cannot be parsed
* [FIX] report an unknown run duration as NaN instead of -1
* [FIX] reject invalid --web.telemetry-path instead of panicking
* [FIX] windows service no longer exits on an unexpected control request
* [FIX] do not log through a nil logger when logger initialisation fails

## 0.1.6 / 2024-06-07

* [FEATURE] go 1.21
* [FEATURE] use metrics resources to check if run is in sucess

## 0.1.5 / 2022-12-15

* [FEATURE] add scrape_error metric for each collector
* [FIX] windows logging to eventlog in service

## 0.1.4 / 2022-12-12

* [FEATURE] support windows event log
* [FIX] puppet config file path on windows 
* [FIX] rename puppet_last_run_catalog_version metric
* [FIX] collect metrics only if access to report file

## 0.1.3 / 2022-12-09

* [FEATURE] support windows service registration 

## 0.1.2 / 2022-12-08

* [FEATURE] add report events/changes metrics

## 0.1.1 / 2022-12-06

* [FEATURE] add windows support
* [FEATURE] add report resources/time metrics
* [FEATURE] add puppet_disabled_lock_info metric
* [FEATURE] add tls from exporter-toolkit and refactoring
* [ENHANCEMENT] add version command and promu for building package
