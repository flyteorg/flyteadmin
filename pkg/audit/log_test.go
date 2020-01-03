package audit

import (
	"context"
	"testing"
	"time"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestLogBuilderLog(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.PrincipalContextKey, "prince")
	tokenIssuedAt := time.Unix(100, 0)
	requestedAt := time.Unix(200, 0)
	sentAt := time.Unix(300, 0)
	err := errors.NewFlyteAdminError(codes.AlreadyExists, "womp womp")
	ctx = context.WithValue(ctx, common.AuditFieldsContextKey, AuthenticatedClientMeta{
		ClientIds:     []string{"12345"},
		TokenIssuedAt: tokenIssuedAt,
		ClientIP:      "192.0.2.1:25",
	})
	builder := NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"my_method", map[string]string{
			"my": "params",
		}, admin.Request_READ_WRITE, requestedAt).WithResponse(sentAt, err)
	assert.EqualValues(t, "Recording request: [{\"principal\":{\"subject\":\"prince\",\"clientId\":\"12345\","+
		"\"tokenIssuedAt\":\"1970-01-01T00:01:40Z\"},\"client\":{\"clientIp\":\"192.0.2.1:25\"},\"request\":"+
		"{\"method\":\"my_method\",\"parameters\":{\"my\":\"params\"},\"mode\":\"READ_WRITE\",\"receivedAt\":"+
		"\"1970-01-01T00:03:20Z\"},\"response\":{\"responseCode\":\"OK\",\"sentAt\":\"1970-01-01T00:05:00Z\"}}]",
		builder.(*logBuilder).formatLogString(context.TODO()))
}
