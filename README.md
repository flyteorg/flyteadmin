Flyteadmin
=============
[![Current Release](https://img.shields.io/github/release/flyteorg/flyteadmin.svg)](https://github.com/flyteorg/flyteadmin/releases/latest)
![Master](https://github.com/flyteorg/flyteadmin/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flyteadmin?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flyteadmin.svg)](https://codecov.io/gh/flyteorg/flyteadmin)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flyteadmin)](https://goreportcard.com/report/github.com/flyteorg/flyteadmin)
![Commit activity](https://img.shields.io/github/commit-activity/w/flyteorg/flyteadmin.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/flyteorg/flyteadmin/latest.svg?style=plastic)

Flyteadmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. Flyteadmin implements the
[AdminService](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/service/admin.proto) which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
Flyteadmin uses a relational style Metadata Store abstracted by [GORM](http://gorm.io/) ORM library.

Before Check-In
---------------

Flyte Admin has a few useful make targets for linting and testing. Please use these before checking in to help suss out
minor bugs and linting errors.

```
  $ make goimports
```

```
  $ make test_unit
```

```
  $ make lint
```
