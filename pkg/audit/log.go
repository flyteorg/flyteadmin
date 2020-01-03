package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type LogBuilder interface {
	WithAuthenticatedCtx(ctx context.Context) LogBuilder
	WithRequest(method string, parameters map[string]string, mode admin.Request_Mode, requestedAt time.Time) LogBuilder
	WithResponse(sentAt time.Time, err error) LogBuilder
	Log(ctx context.Context)
}

var jsonPbMarshaler = jsonpb.Marshaler{}

type logBuilder struct {
	auditLog admin.AuditLog
	readOnly bool
}

func (b *logBuilder) WithAuthenticatedCtx(ctx context.Context) LogBuilder {
	clientMeta := ctx.Value(common.AuditFieldsContextKey)
	switch clientMeta.(type) {
	case AuthenticatedClientMeta:
		tokenIssuedAt, err := ptypes.TimestampProto(clientMeta.(AuthenticatedClientMeta).TokenIssuedAt)
		if err != nil {
			logger.Warningf(ctx, "Failed to convert authenticated TokenIssuedAt to timestamp proto: %v", err)
		}
		b.auditLog.Principal = &admin.Principal{
			Subject:       ctx.Value(common.PrincipalContextKey).(string),
			TokenIssuedAt: tokenIssuedAt,
		}
		if len(clientMeta.(AuthenticatedClientMeta).ClientIds) > 0 {
			b.auditLog.Principal.ClientId = clientMeta.(AuthenticatedClientMeta).ClientIds[0]
		}
		b.auditLog.Client = &admin.Client{
			ClientIp: clientMeta.(AuthenticatedClientMeta).ClientIP,
		}
	default:
		logger.Warningf(ctx, "Failed to parse authenticated client metadata when creating audit log")
	}
	return b
}

// TODO: Also look into passing down HTTP verb
func (b *logBuilder) WithRequest(method string, parameters map[string]string, mode admin.Request_Mode,
	requestedAt time.Time) LogBuilder {
	receivedAt, err := ptypes.TimestampProto(requestedAt)
	if err != nil {
		logger.Warningf(context.TODO(), "Failed to convert authenticated RequestedAt to timestamp proto: %v", err)
	}
	b.auditLog.Request = &admin.Request{
		Method:     method,
		Parameters: parameters,
		Mode:       mode,
		ReceivedAt: receivedAt,
	}
	return b
}

func (b *logBuilder) WithResponse(sentAt time.Time, err error) LogBuilder {
	sentAtProto, err := ptypes.TimestampProto(sentAt)
	if err != nil {
		logger.Warningf(context.TODO(), "Failed to convert authenticated SentAt to timestamp proto: %v", err)
	}

	responseCode := codes.OK.String()
	if err != nil {
		switch err := err.(type) {
		case errors.FlyteAdminError:
			responseCode = err.(errors.FlyteAdminError).Code().String()
		default:
			responseCode = codes.Internal.String()
		}
	}
	b.auditLog.Response = &admin.Response{
		ResponseCode: responseCode,
		SentAt:       sentAtProto,
	}
	return b
}

func (b *logBuilder) formatLogString(ctx context.Context) string {
	auditLog, err := jsonPbMarshaler.MarshalToString(&b.auditLog)
	if err != nil {
		logger.Warningf(ctx, "Failed to marshal audit log to protobuf with err: %v", err)
	}
	return fmt.Sprintf("Recording request: [%s]", auditLog)
}

func (b *logBuilder) Log(ctx context.Context) {
	if b.readOnly {
		logger.Warningf(ctx, "Attempting to record audit log for request: [%+v] more than once. Aborting.", b.auditLog.Request)
	}
	defer func() {
		b.readOnly = true
	}()
	logger.Info(ctx, b.formatLogString(ctx))
}

func NewLogBuilder() LogBuilder {
	return &logBuilder{}
}
