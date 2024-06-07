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
