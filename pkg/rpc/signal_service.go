package rpc

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"

	"github.com/prometheus/client_golang/prometheus"
)

type signalMetrics struct {
	scope        promutils.Scope
	panicCounter prometheus.Counter

	create util.RequestMetrics
	get    util.RequestMetrics
}

type SignalService struct {
	service.UnimplementedSignalServiceServer
	metrics signalMetrics
}

func NewSignalServer(ctx context.Context, adminScope promutils.Scope) *SignalService {
	panicCounter := adminScope.MustNewCounter("initialization_panic",
		"panics encountered initializing the admin service")

	defer func() {
		if err := recover(); err != nil {
			panicCounter.Inc()
			logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	/*databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := adminScope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(adminScope.NewSubScope("errors")), dbScope)
	execCluster := executionCluster.GetExecutionCluster(
		adminScope.NewSubScope("executor").NewSubScope("cluster"),
		kubeConfig,
		master,
		configuration,
		repo)*/

	logger.Info(ctx, "Initializing a new SignalService")
	return &SignalService{
		metrics: signalMetrics{
			scope: adminScope,
			panicCounter: adminScope.MustNewCounter("handler_panic",
				"panics encountered while handling requests to the admin service"),
			create: util.NewRequestMetrics(adminScope, "create_signal"),
			get:    util.NewRequestMetrics(adminScope, "get_signal"),
		},
	}
}

// Intercepts all admin requests to handle panics during execution.
func (s *SignalService) interceptPanic(ctx context.Context, request proto.Message) {
	err := recover()
	if err == nil {
		return
	}

	s.metrics.panicCounter.Inc()
	logger.Fatalf(ctx, "panic-ed for request: [%+v] with err: %v with Stack: %v", request, err, string(debug.Stack()))
}
