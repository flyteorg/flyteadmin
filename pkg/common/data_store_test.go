package common

import (
	"context"
	"testing"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var literalMap = &core.LiteralMap{
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

func TestOffloadLiteralMap(t *testing.T) {
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
		assert.Equal(t, reference.String(), "s3://bucket/metadata/nested/key")
		return nil
	}

	uri, err := OffloadLiteralMap(context.TODO(), mockStorage, literalMap, "nested", "key")
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket/metadata/nested/key", uri.String())
}
