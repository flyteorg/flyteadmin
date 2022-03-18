package dataproxy

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyteadmin/pkg/config"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/graymeta/stow"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type Service struct {
	service.DataProxyServer
	*service.UnimplementedDataProxyServer

	cfg       config.DataProxyConfig
	dataStore *storage.DataStore
}

// CreateUploadLocation creates a temporary signed url to allow callers to upload content.
func (s Service) CreateUploadLocation(ctx context.Context, req *service.CreateUploadLocationRequest) (
	*service.CreateUploadLocationResponse, error) {

	if len(req.Project) == 0 || len(req.Domain) == 0 {
		return nil, fmt.Errorf("prjoect and domain are required parameters")
	}

	if expiresIn := req.ExpiresIn; expiresIn != nil {
		if !expiresIn.IsValid() {
			return nil, fmt.Errorf("expiresIn [%v] is invalid", expiresIn)
		}

		if expiresIn.AsDuration() > s.cfg.Upload.MaxExpiresIn.Duration {
			return nil, fmt.Errorf("expiresIn [%v] cannot exceed max allowed expiration [%v]",
				expiresIn.AsDuration().String(), s.cfg.Upload.MaxExpiresIn.String())
		}
	} else {
		req.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	if len(req.Suffix) == 0 {
		req.Suffix = rand.String(20)
	}

	path := s.dataStore.GetBaseContainerFQN(ctx)
	path, err := s.dataStore.ConstructReference(ctx, path, rand.String(2), req.Project, req.Domain, req.Suffix)
	if err != nil {
		return nil, fmt.Errorf("failed to construct datastore reference. Error: %w", err)
	}

	resp, err := s.dataStore.CreateSignedURL(ctx, path, storage.SignedURLProperties{
		Scope:     stow.ClientMethodPut,
		ExpiresIn: req.ExpiresIn.AsDuration(),
		// TODO: pass max allowed upload size
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create a signed url. Error: %w", err)
	}

	return &service.CreateUploadLocationResponse{
		SignedUrl: resp.URL.String(),
		NativeUrl: path.String(),
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore) Service {
	return Service{
		cfg:       cfg,
		dataStore: dataStore,
	}
}
