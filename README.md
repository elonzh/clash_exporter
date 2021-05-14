# Clash Exporter for Prometheus

[![Docker Pulls](https://img.shields.io/docker/pulls/elonzh/clash_exporter?style=flat-square)](https://hub.docker.com/r/elonzh/clash_exporter)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/elonzh/clash_exporter?style=flat-square)](https://github.com/elonzh/clash_exporter/releases)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/elonzh/clash_exporter/ci?style=flat-square)](https://github.com/elonzh/clash_exporter/actions)
[![codecov](https://img.shields.io/codecov/c/github/elonzh/clash_exporter?style=flat-square&token=w1ngj45JWz)](https://codecov.io/gh/elonzh/clash_exporter)
[![GitHub license](https://img.shields.io/github/license/elonzh/clash_exporter?style=flat-square)](https://github.com/elonzh/clash_exporter/blob/main/LICENSE)

This is a simple server that scrapes [Clash](https://github.com/Dreamacro/clash) stats and exports them via HTTP for
Prometheus consumption.

## Getting Started

To run it:

```bash
./clash_exporter [flags]
```

Help on flags:

```bash
./clash_exporter --help
```

For more information check the [source code documentation][gdocs].

[gdocs]: https://pkg.go.dev/github.com/elonzh/clash_exporter

### TLS and basic authentication

The Clash Exporter supports TLS and basic authentication.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```
