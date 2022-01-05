package validation

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestGetMissingArgumentError(t *testing.T) {
	err := shared.GetMissingArgumentError("foo")
	assert.EqualError(t, err, "missing foo")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateMaxLengthStringField(t *testing.T) {
	err := ValidateMaxLengthStringField("abcdefg", "foo", 6)
	assert.EqualError(t, err, "foo cannot exceed 6 characters")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateMaxMapLengthField(t *testing.T) {
	labels := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	err := ValidateMaxMapLengthField(labels, "foo", 2)
	assert.EqualError(t, err, "foo map cannot exceed 2 entries")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateIdentifier(t *testing.T) {
	err := ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Domain:       "domain",
		Name:         "name",
	}, common.Task)
	assert.EqualError(t, err, "missing project")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Name:         "name",
	}, common.Task)
	assert.EqualError(t, err, "missing domain")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
	}, common.Task)
	assert.EqualError(t, err, "missing name")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
	}, common.Task)
	assert.EqualError(t, err, "unexpected resource type workflow for identifier "+
		"[resource_type:WORKFLOW project:\"project\" domain:\"domain\" ], expected task instead")
}

func TestValidateNamedEntityIdentifierListRequest(t *testing.T) {
	assert.Nil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Domain: "domain",
		Limit:  2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Limit:   2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
	}))
}

func TestValidateVersion(t *testing.T) {
	err := ValidateVersion("")
	assert.EqualError(t, err, "missing version")
}

func TestValidateListTaskRequest(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 10,
	}
	assert.NoError(t, ValidateResourceListRequest(request))
}

func TestValidateListTaskRequest_MissingProject(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: "domain",
			Name:   "name",
		},
		Limit: 10,
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "missing project")
}

func TestValidateListTaskRequest_MissingDomain(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Name:    "name",
		},
		Limit: 10,
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "missing domain")
}

func TestValidateListTaskRequest_MissingName(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Limit: 10,
	}
	assert.NoError(t, ValidateResourceListRequest(request))
}

func TestValidateListTaskRequest_MissingLimit(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "invalid value for limit")
}

func TestValidateParameterMap(t *testing.T) {
	exampleMap := core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Default{
					Default: coreutils.MustMakeLiteral("foo-value"),
				},
			},
		},
	}
	err := validateParameterMap(&exampleMap, "foo")
	assert.NoError(t, err)

	exampleMap = core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: nil, // neither required or defaults
			},
		},
	}
	err = validateParameterMap(&exampleMap, "some text")
	assert.Error(t, err)

	exampleMap = core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			},
		},
	}
	err = validateParameterMap(&exampleMap, "some text")
	assert.NoError(t, err)

	exampleMap = core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Required{
					Required: false,
				},
			},
		},
	}
	err = validateParameterMap(&exampleMap, "some text")
	assert.Error(t, err)
}

func TestValidateToken(t *testing.T) {
	offset, err := ValidateToken("")
	assert.Nil(t, err)
	assert.Equal(t, 0, offset)

	offset, err = ValidateToken("1")
	assert.Nil(t, err)
	assert.Equal(t, 1, offset)

	_, err = ValidateToken("foo")
	assert.NotNil(t, err)

	_, err = ValidateToken("-1")
	assert.NotNil(t, err)
}

func TestValidateActiveLaunchPlanRequest(t *testing.T) {
	err := ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
	)
	assert.Nil(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Domain: "d",
				Name:   "n",
			},
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Name:    "n",
			},
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Domain:  "d",
			},
		},
	)
	assert.Error(t, err)
}

func TestValidateActiveLaunchPlanListRequest(t *testing.T) {
	err := ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
			Domain:  "d",
			Limit:   2,
		},
	)
	assert.Nil(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Domain: "d",
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
			Domain:  "d",
			Limit:   0,
		},
	)
	assert.Error(t, err)
}

func TestValidateOutputData(t *testing.T) {
	t.Run("no output data", func(t *testing.T) {
		assert.NoError(t, ValidateOutputData(nil, 100))
	})
	outputData := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 4,
								},
							},
						},
					},
				},
			},
		},
	}
	t.Run("output data within threshold", func(t *testing.T) {
		assert.NoError(t, ValidateOutputData(outputData, int64(10000000)))
	})
	t.Run("output data greater than threshold", func(t *testing.T) {
		err := ValidateOutputData(outputData, int64(1))
		assert.Equal(t, codes.ResourceExhausted, err.(errors.FlyteAdminError).Code())
	})
}

func TestValidateDatetime(t *testing.T) {
	t.Run("no datetime", func(t *testing.T) {
		assert.NoError(t, ValidateDatetime(""))
	})
	t.Run("datetime with valid format", func(t *testing.T) {
		assert.NoError(t, ValidateDatetime("2021-12-25T15:04:05Z")) // TODO check if this is really the expected format
	})
	t.Run("datetime with invalid format", func(t *testing.T) {
		err := ValidateDatetime("2021-12-25")
		assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
	})
	t.Run("datetime with invalid value", func(t *testing.T) {
		err := ValidateDatetime("1000-00-00T15:04:05.1000000000000Z")
		assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
	})
}
