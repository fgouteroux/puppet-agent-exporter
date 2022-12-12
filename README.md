# puppet-agent-exporter

## Puppet Agent Prometheus Exporter

This [Prometheus](https://prometheus.io/)
[exporter](https://prometheus.io/docs/instrumenting/exporters/)
exposes the status of the Puppet Agent on the host it is running on.

Unlike [puppet-prometheus_reporter](https://github.com/voxpupuli/puppet-prometheus_reporter)
and other solutions that rely on the Puppet Agent successfully a report to the
Puppet Server, this allows you to actively monitor the status of every node in
your environment that is discoverable by Prometheus.

### Metrics Exposed

```
# HELP puppet_agent_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which puppet_agent_exporter was built.
# TYPE puppet_agent_exporter_build_info gauge
puppet_agent_exporter_build_info{branch="test",goversion="go1.19.3",revision="5a65b5769f8394e2d5b034bf28987eaed9da6840",version="0.1.1"} 1
# HELP puppet_config Puppet configuration.
# TYPE puppet_config gauge
puppet_config{environment="sandbox",server="puppetmaster.example.com"} 1
# HELP puppet_disabled_lock_info Puppet state of agent disabled lock.
# TYPE puppet_disabled_lock_info gauge
puppet_disabled_lock_info{disabled_message=""} 0
# HELP puppet_last_run_catalog_version The version of the last attempted Puppet catalog.
# TYPE puppet_last_run_catalog_version gauge
puppet_last_run_catalog_version 1.670338093e+09
# HELP puppet_last_run_at_seconds Time of the last Puppet run.
# TYPE puppet_last_run_at_seconds gauge
puppet_last_run_at_seconds 1.67033806036635e+09
# HELP puppet_last_run_duration_seconds Duration of the last Puppet run.
# TYPE puppet_last_run_duration_seconds gauge
puppet_last_run_duration_seconds 67.197598991
# HELP puppet_last_run_report_changes Changes of the last Puppet run
# TYPE puppet_last_run_report_changes gauge
puppet_last_run_report_changes{type="total"} 0
# HELP puppet_last_run_report_events Events state of the last Puppet run
# TYPE puppet_last_run_report_events gauge
puppet_last_run_report_events{type="failure"} 0
puppet_last_run_report_events{type="success"} 0
puppet_last_run_report_events{type="total"} 0
# HELP puppet_last_run_report_resources Resources state of the last Puppet run
# TYPE puppet_last_run_report_resources gauge
puppet_last_run_report_resources{type="changed"} 1
puppet_last_run_report_resources{type="corrective_change"} 1
puppet_last_run_report_resources{type="failed"} 0
puppet_last_run_report_resources{type="failed_to_restart"} 0
puppet_last_run_report_resources{type="out_of_sync"} 1
puppet_last_run_report_resources{type="restarted"} 0
puppet_last_run_report_resources{type="scheduled"} 0
puppet_last_run_report_resources{type="skipped"} 0
puppet_last_run_report_resources{type="total"} 574
# HELP puppet_last_run_report_time_duration_seconds Resources duration of the last Puppet run.
# TYPE puppet_last_run_report_time_duration_seconds gauge
puppet_last_run_report_time_duration_seconds{type="catalog_application"} 23.69262781366706
puppet_last_run_report_time_duration_seconds{type="config_retrieval"} 13.081214314326644
puppet_last_run_report_time_duration_seconds{type="convert_catalog"} 1.7260215114802122
puppet_last_run_report_time_duration_seconds{type="exec"} 7.371768186000001
puppet_last_run_report_time_duration_seconds{type="fact_generation"} 3.916432438418269
puppet_last_run_report_time_duration_seconds{type="group"} 0.0016724119999999999
puppet_last_run_report_time_duration_seconds{type="node_retrieval"} 2.8758434895426035
puppet_last_run_report_time_duration_seconds{type="package"} 4.396066520999999
puppet_last_run_report_time_duration_seconds{type="plugin_sync"} 21.46335389278829
puppet_last_run_report_time_duration_seconds{type="service"} 0.7416738879999999
puppet_last_run_report_time_duration_seconds{type="transaction_evaluation"} 23.562324536964297
puppet_last_run_report_time_duration_seconds{type="user"} 0.002918061
# HELP puppet_last_run_success 1 if the last Puppet run was successful.
# TYPE puppet_last_run_success gauge
puppet_last_run_success 1
```

### Example Alert Rules

```yaml
groups:
  - name: Puppet
    rules:
      - alert: MultiplePuppetFailing
        expr: count(puppet_last_run_success) - count(puppet_last_run_success == 1) > 4
      - alert: PuppetEnvironmentSet
        expr: puppet_config{environment!=""}
        for: 4h
      - alert: PuppetFailing
        expr: puppet_last_run_success == 0
        for: 40m
      - alert: StalePuppetCatalog
        expr: time() - puppet_last_catalog_version > 3*60*60
        for: 40m
      - alert: LastPuppetTooLongAgo
        expr: time() - puppet_last_run_at_seconds > 3*60*60
        for: 40m
```

We've found it worthwhile to avoid alerting on conditions affecting individual
nodes until that condition has persisted long enough to affect more than one
run. (For example, a transient hiccup in an APT proxy, etc...)

Alerting on the catalog being too old helps catch situations like a node being
set to an environment that no longer exists, or if it's having TLS or network
issues contacting the Puppet Server.

Alerting on a non-default environment being set helps catch operator error,
for example when a node is used to test changes from a branch environment
but forgotten about after that branch is merged.

## TLS and basic authentication

Puppet Agent Exporter supports TLS and basic authentication. This enables better control of the various HTTP endpoints.

To use TLS and/or basic authentication, you need to pass a configuration file using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

## Project Status: **Works For Us**

This is an open-source release of something we've used for quite a while
internally.

*   The package APIs functions are likely to be refactored drastically,
    potentially without warning in release notes.

    (This isn't really meant to be a library.)

*   We may add a systemd unit to the packaging at some point.

*   We probably won't make breaking changes to the arguments without warning.

### Areas needing improvement

*   Test coverage
*   General code hygiene and refactoring
*   Better packaging (supporting RPM distros, including a systemd unit, etc...)

## Contributing

Contributions considered, but be aware that this is mostly just something we
needed. It's public because there's no reason anyone else should have to waste
an afternoon (or more) building something similar, and we think the approach
is good enough that others would benefit from adopting.

This project is licensed under the [Apache License, Version 2.0](LICENSE).

Please include a `Signed-off-by` in all commits, per
[Developer Certificate of Origin version 1.1](DCO).
