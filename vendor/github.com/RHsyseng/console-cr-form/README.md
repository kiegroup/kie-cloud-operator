# Custom Resource Form Generation

[![Go Report Card](https://goreportcard.com/badge/github.com/RHsyseng/console-cr-form)](https://goreportcard.com/report/github.com/RHsyseng/console-cr-form)
[![Build Status](https://travis-ci.org/RHsyseng/console-cr-form.svg?branch=master)](https://travis-ci.org/RHsyseng/console-cr-form)

## Requirements

- go v1.10+
- dep v0.5.0+
- npm

## Build

### Clean generated Go files
```bash
make clean
```

### Build a local binary
```bash
make
```

### Only rebuild npm modules / webpack
```bash
make npm
```

## Run & Test
```bash
./build/console-cr-form
```
