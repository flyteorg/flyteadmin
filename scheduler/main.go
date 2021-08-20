package main

import (
	"github.com/flyteorg/flyteadmin/scheduler/entrypoints"
	"github.com/golang/glog"
)

func main() {
	glog.V(2).Info("Beginning Flyte Scheduler")
	err := entrypoints.Execute()
	if err != nil {
		panic(err)
	}
}
