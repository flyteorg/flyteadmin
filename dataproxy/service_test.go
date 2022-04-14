package dataproxy

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
	storageMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	s, err := NewService(config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{},
	}, dataStore)
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func init() {
	labeled.SetMetricKeys(contextutils.DomainKey)
}

func Test_createShardedStorageLocation(t *testing.T) {
	selector, err := ioutils.NewBase36PrefixShardSelector(context.TODO())
	assert.NoError(t, err)
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	loc, err := createShardedStorageLocation(context.Background(), selector, dataStore, config.DataProxyUploadConfig{
		StoragePrefix: "blah",
	})
	assert.NoError(t, err)
	assert.Equal(t, "/u8/blah", loc.String())
}

func TestCreateUploadLocation(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	s, err := NewService(config.DataProxyConfig{}, dataStore)
	assert.NoError(t, err)
	t.Run("No project/domain", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{})
		assert.Error(t, err)
	})

	t.Run("unsupported operation by InMemory DataStore", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project: "hello",
			Domain:  "world",
		})
		assert.Error(t, err)
	})

	t.Run("Invalid expiry", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project:   "hello",
			Domain:    "world",
			ExpiresIn: durationpb.New(-time.Hour),
		})
		assert.Error(t, err)
	})

	t.Run("jfkdls", func(t *testing.T) {
		hexString := "04977c0f4640305dcc9f6fff542c7f09"
		data, err := hex.DecodeString(hexString)
		assert.NoError(t, err)

		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project:    "hello",
			Domain:     "world",
			ExpiresIn:  durationpb.New(0),
			ContentMd5: data,
		})
		fmt.Println(err)
		assert.Error(t, err)
	})
}

func TestCreafteUploadLocation(t *testing.T) {
	mockClient := storageMocks.GetMockStorageClient()
	s, err := NewService(config.DataProxyConfig{}, mockClient)
	assert.NoError(t, err)

	t.Run("jfkdls", func(t *testing.T) {
		hexString := "04977c0f4640305dcc9f6fff542c7f09"
		data, err := hex.DecodeString(hexString)
		assert.NoError(t, err)


		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project:    "hello",
			Domain:     "world",
			ExpiresIn:  durationpb.New(0),
			ContentMd5: data,
		})
		fmt.Println(err)
		assert.Error(t, err)
	})
}


//func TestJdls(t *testing.T) {
//	hexString := "04977c0f4640305dcc9f6fff542c7f09"
//	data, err := hex.DecodeString(hexString)
//	assert.NoError(t, err)
//
//	md5 := base64.StdEncoding.EncodeToString(data)
//	urlSafeMd5 := base32.StdEncoding.EncodeToString(data)
//	fmt.Println(md5)
//	fmt.Println(urlSafeMd5)
//}
