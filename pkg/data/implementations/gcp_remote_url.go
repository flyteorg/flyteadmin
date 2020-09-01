package implementations

import (
	"context"
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"golang.org/x/oauth2"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	gcs "cloud.google.com/go/storage"
	"github.com/lyft/flyteadmin/pkg/data/interfaces"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
	"google.golang.org/api/option"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
	"google.golang.org/grpc/codes"
)

const gcsScheme = "gs"

// GCP-specific implementation of RemoteURLInterface
type GCPRemoteURL struct {
	gcsClient        *gcs.Client
	signDuration     time.Duration
	signingPrincipal string
}

type GCPGCSObject struct {
	bucket string
	object string
}

func (g *GCPRemoteURL) splitURI(ctx context.Context, uri string) (GCPGCSObject, error) {
	scheme, container, key, err := storage.DataReference(uri).Split()
	if err != nil {
		return GCPGCSObject{}, err
	}
	if scheme != gcsScheme {
		logger.Debugf(ctx, "encountered unexpected scheme: %s for GCS URI: %s", scheme, uri)
		return GCPGCSObject{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"unexpected scheme %s for GCS URI", scheme)
	}
	return GCPGCSObject{
		bucket: container,
		object: key,
	}, nil
}

func (g *GCPRemoteURL) signURL(ctx context.Context, gcsURI GCPGCSObject) (string, error) {
	c, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		return "", err
	}

	opts := &gcs.SignedURLOptions{
		Method:         "GET",
		GoogleAccessID: g.signingPrincipal,
		SignBytes: func(b []byte) ([]byte, error) {
			req := &credentialspb.SignBlobRequest{
				Payload: b,
				Name:    "projects/-/serviceAccounts/" + g.signingPrincipal,
			}
			resp, err := c.SignBlob(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.SignedBlob, nil
		},
		Expires: time.Now().Add(g.signDuration),
	}

	return gcs.SignedURL(gcsURI.bucket, gcsURI.object, opts)
}

func (g *GCPRemoteURL) Get(ctx context.Context, uri string) (admin.UrlBlob, error) {
	logger.Debugf(ctx, "Getting signed url for - %s", uri)
	gcsURI, err := g.splitURI(ctx, uri)
	if err != nil {
		logger.Debugf(ctx, "failed to extract gcs bucket and object from uri: %s", uri)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid uri: %s", uri)
	}

	// First, get the size of the url blob.
	attrs, err := g.gcsClient.Bucket(gcsURI.bucket).Object(gcsURI.object).Attrs(ctx)
	if err != nil {
		logger.Debugf(ctx, "failed to get object size for %s with %v", uri, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to get object size for %s with %v", uri, err)
	}

	// The second return argument here is the GetObjectOutput, which we don't use below.
	urlStr, err := g.signURL(ctx, gcsURI)
	if err != nil {
		logger.Warning(ctx,
			"failed to presign url for uri [%s] for %v with err %v", uri, g.signDuration, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to presign url for uri [%s] for %v with err %v", uri, g.signDuration, err)
	}
	return admin.UrlBlob{
		Url:   urlStr,
		Bytes: attrs.Size,
	}, nil
}

type impersonatedTokenSource struct {
	signingPrincipal string
}

func (ts impersonatedTokenSource) Token() (*oauth2.Token, error) {
	c, err := credentials.NewIamCredentialsClient(context.Background())
	if err != nil {
		return nil, err
	}

	req := credentialspb.GenerateAccessTokenRequest{
		Name:  "projects/-/serviceAccounts/" + ts.signingPrincipal,
		Scope: []string{"https://www.googleapis.com/auth/devstorage.read_only"},
	}

	resp, err := c.GenerateAccessToken(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: resp.AccessToken,
		Expiry:      resp.ExpireTime.AsTime(),
	}, nil
}

func NewGCPRemoteURL(signingPrincipal string, signDuration time.Duration) interfaces.RemoteURLInterface {
	gcsClient, err := gcs.NewClient(context.Background(), option.WithScopes(gcs.ScopeReadOnly), option.WithTokenSource(impersonatedTokenSource{signingPrincipal: signingPrincipal}))
	if err != nil {
		panic(err)
	}

	return &GCPRemoteURL{
		gcsClient:        gcsClient,
		signDuration:     signDuration,
		signingPrincipal: signingPrincipal,
	}
}
