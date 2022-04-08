package dataproxy

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/duration"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyteadmin/pkg/config"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/flyteorg/stow"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type Service struct {
	service.DataProxyServiceServer

	cfg           config.DataProxyConfig
	dataStore     *storage.DataStore
	shardSelector ioutils.ShardSelector
}

// CreateUploadLocationBatch creates a signed url to upload artifacts to for a given project/domain.
func (s Service) CreateUploadLocationBatch(ctx context.Context, req *service.CreateUploadLocationBatchRequest) (*service.CreateUploadLocationBatchResponse, error) {
	if len(req.Project) == 0 || len(req.Domain) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project and domain are required parameters")
	}

	if len(req.Items) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least one item is required")
	}

	if len(req.Items) > s.cfg.Upload.MaxBatchSize {
		return nil, status.Errorf(codes.InvalidArgument, "batch size [%v] exceeded the allowed size [%v]",
			len(req.Items), s.cfg.Upload.MaxBatchSize)
	}

	if expiresIn := req.ExpiresIn; expiresIn != nil {
		if err := validateExpiresIn(expiresIn, s.cfg.Upload.MaxExpiresIn.Duration); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "expiresIn [%v] failed validation. Error: %v", expiresIn, err)
		}
	} else {
		req.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	if len(req.Suffix) == 0 {
		req.Suffix = rand.String(s.cfg.Upload.DefaultFileNameLength)
	}

	items := make([]*service.ItemUploadInfo, 0, len(req.Items))
	for _, item := range req.Items {
		md5Base64 := base64.StdEncoding.EncodeToString(item.ContentMd5)
		storagePath, err := createShardedStorageLocation(ctx, s.shardSelector, s.dataStore, s.cfg.Upload, req.Project,
			req.Domain, req.Suffix, md5Base64)
		if err != nil {
			return nil, err
		}

		resp, err := s.dataStore.CreateSignedURL(ctx, storagePath, storage.SignedURLProperties{
			Scope:      stow.ClientMethodPut,
			ExpiresIn:  req.ExpiresIn.AsDuration(),
			ContentMD5: md5Base64,
			// TODO: pass max allowed upload size
		})

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to create a signed url. Error: %v", err)
		}

		items = append(items, &service.ItemUploadInfo{
			SignedUrl: resp.URL.String(),
			NativeUrl: storagePath.String(),
		})
	}

	return &service.CreateUploadLocationBatchResponse{
		Items:     items,
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

func validateExpiresIn(expiresIn *duration.Duration, maxExpiresIn time.Duration) error {
	if expiresIn != nil {
		if !expiresIn.IsValid() {
			return fmt.Errorf("expiresIn [%v] is invalid", expiresIn)
		}

		if expiresIn.AsDuration() > maxExpiresIn {
			return fmt.Errorf("expiresIn [%v] cannot exceed max allowed expiration [%v]",
				expiresIn.AsDuration().String(), maxExpiresIn.String())
		}
	}

	return nil
}

// CreateUploadLocation creates a temporary signed url to allow callers to upload content.
func (s Service) CreateUploadLocation(ctx context.Context, req *service.CreateUploadLocationRequest) (
	*service.CreateUploadLocationResponse, error) {

	if len(req.Project) == 0 || len(req.Domain) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "prjoect and domain are required parameters")
	}

	if len(req.ContentMd5) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "content_md5 is a required parameter")
	}

	if expiresIn := req.ExpiresIn; expiresIn != nil {
		if err := validateExpiresIn(expiresIn, s.cfg.Upload.MaxExpiresIn.Duration); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "expiresIn [%v] failed validation. Error: %v", expiresIn, err)
		}
	} else {
		req.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	if len(req.Suffix) == 0 {
		req.Suffix = rand.String(s.cfg.Upload.DefaultFileNameLength)
	}

	storagePath, err := createShardedStorageLocation(ctx, s.shardSelector, s.dataStore, s.cfg.Upload, req.Project,
		req.Domain, req.Suffix, req.ContentMd5)
	if err != nil {
		return nil, err
	}

	resp, err := s.dataStore.CreateSignedURL(ctx, storagePath, storage.SignedURLProperties{
		Scope:      stow.ClientMethodPut,
		ExpiresIn:  req.ExpiresIn.AsDuration(),
		ContentMD5: req.ContentMd5,
		// TODO: pass max allowed upload size
	})

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create a signed url. Error: %v", err)
	}

	return &service.CreateUploadLocationResponse{
		SignedUrl: resp.URL.String(),
		NativeUrl: storagePath.String(),
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

// createShardedStorageLocation creates a location in storage destination to maximize read/write performance in most
// block stores. The final location should look something like: s3://<my bucket>/<shard length>/<file name>
func createShardedStorageLocation(ctx context.Context, shardSelector ioutils.ShardSelector, store *storage.DataStore,
	cfg config.DataProxyUploadConfig, keyParts ...string) (storage.DataReference, error) {
	keySuffixArr := make([]string, 0, 4)
	if len(cfg.StoragePrefix) > 0 {
		keySuffixArr = append(keySuffixArr, cfg.StoragePrefix)
	}

	keySuffixArr = append(keySuffixArr, keyParts...)
	prefix, err := shardSelector.GetShardPrefix(ctx, []byte(strings.Join(keySuffixArr, "/")))
	if err != nil {
		return "", err
	}

	storagePath, err := store.ConstructReference(ctx, store.GetBaseContainerFQN(ctx),
		append([]string{prefix}, keySuffixArr...)...)
	if err != nil {
		return "", fmt.Errorf("failed to construct datastore reference. Error: %w", err)
	}

	return storagePath, nil
}

func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore) (Service, error) {
	// Context is not used in the constructor. Should ideally be removed.
	selector, err := ioutils.NewBase36PrefixShardSelector(context.TODO())
	if err != nil {
		return Service{}, err
	}

	return Service{
		cfg:           cfg,
		dataStore:     dataStore,
		shardSelector: selector,
	}, nil
}
