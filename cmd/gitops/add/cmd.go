package add

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add/app"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add/clusters"
)

func GetCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new Weave GitOps resource",
		Example: `
# Add an application to gitops from local git repository
gitops add app . --name <app-name>

# Add a new cluster using a CAPI template
gitops add cluster`,
	}

	cmd.AddCommand(clusters.ClusterCommand(endpoint, client))
	cmd.AddCommand(app.Cmd)

	return cmd
}
