package auth

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/auth/interfaces"
)

var (
	emptyIdentityContext = IdentityContext{}
)

type IdentityContext struct {
	audience        string
	userID          string
	appID           string
	authenticatedAt time.Time
	userInfo        interfaces.UserInfo
}

func (c IdentityContext) Audience() string {
	return c.audience
}

func (c IdentityContext) UserID() string {
	return c.userID
}

func (c IdentityContext) AppID() string {
	return c.appID
}

func (c IdentityContext) UserInfo() interfaces.UserInfo {
	return c.userInfo
}

func (c IdentityContext) IsEmpty() bool {
	return c == emptyIdentityContext
}

func (c IdentityContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyIdentityContext, c)
}

func (c IdentityContext) AuthenticatedAt() time.Time {
	return c.authenticatedAt
}

func NewIdentityContext(audience, userID, appID string, authenticatedAt time.Time, userInfo interfaces.UserInfo) IdentityContext {
	return IdentityContext{
		audience:        audience,
		userID:          userID,
		appID:           appID,
		userInfo:        userInfo,
		authenticatedAt: authenticatedAt,
	}
}

func IdentityContextFromContext(ctx context.Context) IdentityContext {
	existing := ctx.Value(ContextKeyIdentityContext)
	if existing != nil {
		return existing.(IdentityContext)
	}

	return emptyIdentityContext
}
