package dataproxy

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

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

	cfg                  config.DataProxyConfig
	dataStore            *storage.DataStore
	shardSelector        ioutils.ShardSelector
	nodeExecutionManager interfaces.NodeExecutionInterface
	taskExecutionManager interfaces.TaskExecutionInterface
}

// CreateUploadLocation creates a temporary signed url to allow callers to upload content.
func (s Service) CreateUploadLocation(ctx context.Context, req *service.CreateUploadLocationRequest) (
	*service.CreateUploadLocationResponse, error) {

	if len(req.Project) == 0 || len(req.Domain) == 0 {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "project and domain are required parameters")
	}

	if len(req.ContentMd5) == 0 {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "content_md5 is a required parameter")
	}

	if expiresIn := req.ExpiresIn; expiresIn != nil {
		if !expiresIn.IsValid() {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "expiresIn [%v] is invalid", expiresIn)
		}

		if expiresIn.AsDuration() > s.cfg.Upload.MaxExpiresIn.Duration {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "expiresIn [%v] cannot exceed max allowed expiration [%v]",
				expiresIn.AsDuration().String(), s.cfg.Upload.MaxExpiresIn.String())
		}
	} else {
		req.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	if len(req.Filename) == 0 {
		req.Filename = rand.String(s.cfg.Upload.DefaultFileNameLength)
	}

	md5 := base64.StdEncoding.EncodeToString(req.ContentMd5)
	urlSafeMd5 := base32.StdEncoding.EncodeToString(req.ContentMd5)

	storagePath, err := createStorageLocation(ctx, s.dataStore, s.cfg.Upload,
		req.Project, req.Domain, urlSafeMd5, req.Filename)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create shardedStorageLocation, Error: %v", err)
	}

	resp, err := s.dataStore.CreateSignedURL(ctx, storagePath, storage.SignedURLProperties{
		Scope:      stow.ClientMethodPut,
		ExpiresIn:  req.ExpiresIn.AsDuration(),
		ContentMD5: md5,
	})

	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create a signed url. Error: %v", err)
	}

	return &service.CreateUploadLocationResponse{
		SignedUrl: resp.URL.String(),
		NativeUrl: storagePath.String(),
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

// CreateDownloadLink retrieves the requested artifact type for a given execution (wf, node, task) as a signed url(s).
func (s Service) CreateDownloadLink(ctx context.Context, req *service.CreateDownloadLinkRequest) (
	resp *service.CreateDownloadLinkResponse, err error) {
	if req, err = s.validateCreateDownloadLinkRequest(req); err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "error while validating request. Error: %v", err)
	}

	// Lookup task, node, workflow execution
	var nativeURL string
	if nodeExecutionIDEnvelope, casted := req.GetSource().(*service.CreateDownloadLinkRequest_NodeExecutionId); casted {
		node, err := s.nodeExecutionManager.GetNodeExecution(ctx, admin.NodeExecutionGetRequest{
			Id: nodeExecutionIDEnvelope.NodeExecutionId,
		})

		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to find node execution [%v]. Error: %v", nodeExecutionIDEnvelope.NodeExecutionId, err)
		}

		switch req.GetArtifactType() {
		case service.ArtifactType_ARTIFACT_TYPE_DECK:
			nativeURL = node.Closure.DeckUri
		}
	} else {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unsupported source [%v]", reflect.TypeOf(req.GetSource()))
	}

	if len(nativeURL) == 0 {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "no deckUrl found for request [%+v]", req)
	}

	signedURLResp, err := s.dataStore.CreateSignedURL(ctx, storage.DataReference(nativeURL), storage.SignedURLProperties{
		Scope:     stow.ClientMethodGet,
		ExpiresIn: req.ExpiresIn.AsDuration(),
	})

	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create a signed url. Error: %v", err)
	}

	return &service.CreateDownloadLinkResponse{
		SignedUrl: []string{signedURLResp.URL.String()},
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

// CreateDownloadLocation creates a temporary signed url to allow callers to download content.
func (s Service) CreateDownloadLocation(ctx context.Context, req *service.CreateDownloadLocationRequest) (
	*service.CreateDownloadLocationResponse, error) {

	if err := s.validateCreateDownloadLocationRequest(req); err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "error while validating request: %v", err)
	}

	resp, err := s.dataStore.CreateSignedURL(ctx, storage.DataReference(req.NativeUrl), storage.SignedURLProperties{
		Scope:     stow.ClientMethodGet,
		ExpiresIn: req.ExpiresIn.AsDuration(),
	})

	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create a signed url. Error: %v", err)
	}

	return &service.CreateDownloadLocationResponse{
		SignedUrl: resp.URL.String(),
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

func (s Service) validateCreateDownloadLocationRequest(req *service.CreateDownloadLocationRequest) error {
	validatedExpiresIn, err := validateDuration(req.ExpiresIn, s.cfg.Download.MaxExpiresIn.Duration)
	if err != nil {
		return fmt.Errorf("expiresIn is invalid. Error: %w", err)
	}

	req.ExpiresIn = validatedExpiresIn

	if _, err := url.Parse(req.NativeUrl); err != nil {
		return fmt.Errorf("failed to parse native_url [%v]",
			req.NativeUrl)
	}

	return nil
}

func validateDuration(input *durationpb.Duration, maxAllowed time.Duration) (*durationpb.Duration, error) {
	if input == nil {
		return durationpb.New(maxAllowed), nil
	}

	if !input.IsValid() {
		return nil, fmt.Errorf("input duration [%v] is invalid", input)
	}

	if input.AsDuration() < 0 {
		return nil, fmt.Errorf("input duration [%v] should not less than 0",
			input.AsDuration().String())
	} else if input.AsDuration() > maxAllowed {
		return nil, fmt.Errorf("input duration [%v] cannot exceed max allowed expiration [%v]",
			input.AsDuration(), maxAllowed)
	}

	return input, nil
}

func (s Service) validateCreateDownloadLinkRequest(req *service.CreateDownloadLinkRequest) (*service.CreateDownloadLinkRequest, error) {
	validatedExpiresIn, err := validateDuration(req.ExpiresIn, s.cfg.Download.MaxExpiresIn.Duration)
	if err != nil {
		return nil, fmt.Errorf("expiresIn is invalid. Error: %w", err)
	}

	req.ExpiresIn = validatedExpiresIn

	if req.GetArtifactType() == service.ArtifactType_ARTIFACT_TYPE_UNDEFINED {
		return nil, fmt.Errorf("invalid artifact type [%v]", req.GetArtifactType())
	}

	if req.GetSource() == nil {
		return nil, fmt.Errorf("source is required. Provided nil")
	}

	return req, nil
}

// createStorageLocation creates a location in storage destination to maximize read/write performance in most
// block stores. The final location should look something like: s3://<my bucket>/<file name>
func createStorageLocation(ctx context.Context, store *storage.DataStore,
	cfg config.DataProxyUploadConfig, keyParts ...string) (storage.DataReference, error) {

	storagePath, err := store.ConstructReference(ctx, store.GetBaseContainerFQN(ctx),
		append([]string{cfg.StoragePrefix}, keyParts...)...)
	if err != nil {
		return "", fmt.Errorf("failed to construct datastore reference. Error: %w", err)
	}

	return storagePath, nil
}

func (s Service) validateResolveArtifactRequest(req *service.ResolveArtifactRequest) error {
	if req.GetFlyteUrl() == "" {
		return fmt.Errorf("source is required. Provided empty string")
	}
	if !strings.HasPrefix(req.GetFlyteUrl(), "flyte://") {
		return fmt.Errorf("request does not start with the correct prefix")
	}

	return nil
}

type IOType int

const (
	INPUT = iota
	OUTPUT
	DECK
)

func ParseFlyteUrl(flyteUrl string) (core.NodeExecutionIdentifier, *int, IOType, error) {
	// flyteUrl is of the form flyte://v1/project/domain/execution_id/node_id/attempt/[iod]
	// where i stands for inputs.pb o for outputs.pb and d for the flyte deck
	// If the retry attempt is missing, the io requested is assumed to be for the node instead of the task execution
	zero := 0
	re, err := regexp.Compile("flyte://v1/([iod])/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)(?:/([0-9]+))?")
	if err != nil {
		return core.NodeExecutionIdentifier{}, &zero, 0, err
	}
	re.MatchString(flyteUrl)
	matches := re.FindStringSubmatch(flyteUrl)
	if len(matches) != 7 && len(matches) != 6 {
		return core.NodeExecutionIdentifier{}, &zero, 0, fmt.Errorf("failed to parse flyte url, only %d matches found", len(matches))
	}
	proj := matches[2]
	domain := matches[3]
	executionId := matches[4]
	nodeId := matches[5]
	var attempt *int // nil means node execution, not a task execution
	if len(matches) == 7 && matches[6] != "" {
		a, err := strconv.Atoi(matches[6])
		if err != nil {
			return core.NodeExecutionIdentifier{}, &zero, 0, fmt.Errorf("failed to parse attempt, %s", err)
		}
		attempt = &a
	}
	var ioType IOType
	switch matches[1] {
	case "i":
		ioType = INPUT
	case "o":
		ioType = OUTPUT
	case "d":
		ioType = DECK
	}

	return core.NodeExecutionIdentifier{
		NodeId: nodeId,
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: proj,
			Domain:  domain,
			Name:    executionId,
		},
	}, attempt, ioType, nil
}

// ResolveArtifact tries to return the raw remote URL. In cases where only the raw data is available, we will return
// an error code that the frontend (flytekit) should know how to handle.
func (s Service) ResolveArtifact(ctx context.Context, req *service.ResolveArtifactRequest) (
	*service.ResolveArtifactResponse, error) {

	logger.Debugf(ctx, "resolving flyte url query: %s", req.GetFlyteUrl())
	var resolvedURL string
	err := s.validateResolveArtifactRequest(req)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to validate resolve artifact request. Error: %v", err)
	}

	// Get the node execution id and other information
	nodeExecId, attempt, ioType, err := ParseFlyteUrl(req.GetFlyteUrl())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to parse artifact url Error: %v", err)
	}

	logger.Debugf(ctx, "resolved to node exec id %s, attempt %v, type %d", nodeExecId, attempt, ioType)
	// always get the node execution
	node, err := s.nodeExecutionManager.GetNodeExecution(ctx, admin.NodeExecutionGetRequest{
		Id: &nodeExecId,
	})
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to find node execution [%v]. Error: %v", nodeExecId, err)
	}
	if attempt == nil || ioType == DECK {
		// get the node execution io link, if available.
		if ioType == INPUT {
			resolvedURL = node.InputUri
		} else if ioType == OUTPUT {
			// todo: why is this deprecated, this is what the get data endpoint uses.
			resolvedURL = node.Closure.GetOutputUri()
		} else if ioType == DECK {
			// todo: why is there no deck uri for task closure?
			resolvedURL = node.Closure.DeckUri
		}
	} else {
		taskExecs, err := s.taskExecutionManager.ListTaskExecutions(ctx, admin.TaskExecutionListRequest{
			NodeExecutionId: &nodeExecId,
			Limit:           1,
			Filters:         fmt.Sprintf("eq(retry_attempt,%s)", strconv.Itoa(*attempt)),
		})
		if err != nil || len(taskExecs.TaskExecutions) == 0 {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to list task executions [%v]. Error: %v", nodeExecId, err)
		}
		taskExec := taskExecs.TaskExecutions[0]
		if ioType == INPUT {
			resolvedURL = taskExec.InputUri
		} else if ioType == OUTPUT {
			resolvedURL = taskExec.Closure.GetOutputUri()
		}
	}
	if resolvedURL == "" {
		return nil, errors.NewFlyteAdminErrorf(codes.NotFound, "failed to resolve [%s]. Error: %v", req.GetFlyteUrl(), err)
	}

	return &service.ResolveArtifactResponse{
		NativeUrl: resolvedURL,
	}, nil
}

func NewService(cfg config.DataProxyConfig,
	nodeExec interfaces.NodeExecutionInterface,
	dataStore *storage.DataStore,
	taskExec interfaces.TaskExecutionInterface) (Service, error) {

	// Context is not used in the constructor. Should ideally be removed.
	selector, err := ioutils.NewBase36PrefixShardSelector(context.TODO())
	if err != nil {
		return Service{}, err
	}

	return Service{
		cfg:                  cfg,
		dataStore:            dataStore,
		shardSelector:        selector,
		nodeExecutionManager: nodeExec,
		taskExecutionManager: taskExec,
	}, nil
}
