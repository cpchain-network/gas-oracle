# gas-oracle


<!--
parent:
  order: false
-->

<div align="center">
  <h1> gas-oracle repo </h1>
</div>

<div align="center">
  <a href="https://github.com/cpchain-network/gas-oracle/releases/latest">
    <img alt="Version" src="https://img.shields.io/github/tag/cpchain-network/gas-oracle.svg" />
  </a>
  <a href="https://github.com/cpchain-network/gas-oracle/blob/main/LICENSE">
    <img alt="License: Apache-2.0" src="https://img.shields.io/github/license/cpchain-network/gas-oracle.svg" />
  </a>
  <a href="https://pkg.go.dev/github.com/cpchain-network/gas-oracle">
    <img alt="GoDoc" src="https://godoc.org/github.com/cpchain-network/gas-oracle?status.svg" />
  </a>
</div>

gas-oracle is project which can sync any evm chain gas fee with a grpc service for cpchain bridge.

**Tips**:
- need [Go 1.23+](https://golang.org/dl/)
- need [Postgresql](https://www.postgresql.org/)


## Install

### Install dependencies
```bash
go mod tidy
```
### build
```bash
cd gas-oracle
make
```

### Config env

- yaml config, you can [gas-oracle.toml](https://github.com/cpchain-network/gas-oracle/blob/main/gas-oracle.yaml) file and config your real env value.

### start index
```bash
./gas-oracle index -c ./gas-oracle.yaml
```

### start grpc
```bash
./gas-oracle grpc -c ./gas-oracle.yaml
```

## Contribute

TBD
