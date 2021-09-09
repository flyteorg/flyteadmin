package entrypoints

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	adminClient "github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var preCheckRunCmd = &cobra.Command{
	Use:   "precheck",
	Short: "This command will check pre requirement for scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		err := retry.Do(
			func() error {
				config := adminClient.GetConfig(ctx)
				host := strings.Split(config.Endpoint.Host, ":")
				response, err := http.Get(fmt.Sprintf("http://%s/healthcheck", host))
				if err != nil {
					return err
				}
				if response.StatusCode == 200 {
					return nil
				}
				return errors.New("something goes wrong")
			},
			retry.Delay(30*time.Second),
			retry.Attempts(100),
		)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(preCheckRunCmd)
}
