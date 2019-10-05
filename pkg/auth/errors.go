package auth

import (
	"github.com/lyft/flytestdlib/errors"
)

const (
	ErrConfigFileRead errors.ErrorCode = "CONFIG_OPTION_FILE_READ_FAILED"
	ErrTokenNil                        = "EMPTY_OAUTH_TOKEN"
)
