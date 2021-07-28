package entrypoints

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyteorg/flytestdlib/logger"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flyteorg/flyteadmin/auth"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

const (
	PodNamespaceEnvVar  = "POD_NAMESPACE"
	podDefaultNamespace = "default"
)

var (
	secretName       string
	secretsLocalPath string
	forceUpdate      bool
)

var secretsCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets"},
}

var secretsPersistCmd = &cobra.Command{
	Use: "create",
	Long: `Creates a new secret (or noop if one exists unless --force is provided) using keys found in the provided path.
If POD_NAMESPACE env var is set, the secret will be created in that namespace.
`,
	Example: `
Create a secret using default name (flyte-admin-auth) in default namespace
flyteadmin secret create --fromPath=/path/in/container

Override an existing secret if one exists (reads secrets from default path /etc/secrets/):
flyteadmin secret create --name "my-auth-secrets" --force
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return persistSecrets(context.Background(), cmd.Flags())
	},
}

func init() {
	secretsPersistCmd.Flags().StringVar(&secretName, "name", "flyte-admin-auth", "Chooses secret name to create/update")
	secretsPersistCmd.Flags().StringVar(&secretsLocalPath, "fromPath", filepath.Join(string(os.PathSeparator), "etc", "secrets"), "Chooses secret name to create/update")
	secretsPersistCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Whether to update the secret if one exists")
	secretsCmd.AddCommand(secretsPersistCmd)
	secretsCmd.AddCommand(auth.GetInitSecretsCommand())

	RootCmd.AddCommand(secretsCmd)
}

func buildK8sSecretData(_ context.Context, localPath string) (map[string][]byte, error) {
	secretsData := make(map[string][]byte, 4)

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		secretsData[strings.TrimPrefix(path, filepath.Dir(path)+string(filepath.Separator))] = data
		return nil
	})

	if err != nil {
		return nil, err
	}

	return secretsData, nil
}

func persistSecrets(ctx context.Context, _ *pflag.FlagSet) error {
	serverCfg := config.GetConfig()
	configuration := runtime.NewConfigurationProvider()
	scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope)
	clusterClient, err := impl.NewInCluster(scope.NewSubScope("secrets"), serverCfg.KubeConfig, serverCfg.Master)
	if err != nil {
		return err
	}

	targets := clusterClient.GetAllValidTargets()
	// Since we are targeting the cluster Admin is running in, this list should contain exactly one item
	if len(targets) != 1 {
		return fmt.Errorf("expected exactly 1 valid target cluster. Found [%v]", len(targets))
	}

	clusterCfg := targets[0].Config
	kubeClient, err := kubernetes.NewForConfig(&clusterCfg)
	if err != nil {
		return errors.Wrapf("INIT", err, "Error building kubernetes clientset")
	}

	podNamespace, found := os.LookupEnv(PodNamespaceEnvVar)
	if !found {
		podNamespace = podDefaultNamespace
	}

	secretsData, err := buildK8sSecretData(ctx, secretsLocalPath)
	if err != nil {
		return errors.Wrapf("INIT", err, "Error building k8s secret's data field.")
	}

	secretsClient := kubeClient.CoreV1().Secrets(podNamespace)
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: podNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretsData,
	}

	_, err = secretsClient.Create(ctx, newSecret, metav1.CreateOptions{})

	if err != nil && kubeErrors.IsAlreadyExists(err) {
		if forceUpdate {
			logger.Infof(ctx, "A secret already exists with the same name. Attempting to update it.")
			_, err = secretsClient.Update(ctx, newSecret, metav1.UpdateOptions{})
		} else {
			var existingSecret *corev1.Secret
			existingSecret, err = secretsClient.Get(ctx, newSecret.Name, metav1.GetOptions{})
			if err != nil {
				logger.Infof(ctx, "Failed to retrieve existing secret. Error: %v", err)
				return err
			}

			if existingSecret.Data == nil {
				existingSecret.Data = map[string][]byte{}
			}

			needsUpdate := false
			for key, val := range secretsData {
				if _, found := existingSecret.Data[key]; !found {
					existingSecret.Data[key] = val
					needsUpdate = true
				}
			}

			if needsUpdate {
				_, err = secretsClient.Update(ctx, existingSecret, metav1.UpdateOptions{})
				if err != nil && kubeErrors.IsConflict(err) {
					logger.Infof(ctx, "Another instance of flyteadmin has updated the same secret. Ignoring this update")
					err = nil
				}
			}
		}

		return err
	}

	return err
}
