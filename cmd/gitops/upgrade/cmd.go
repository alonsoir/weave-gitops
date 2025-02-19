package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

var upgradeCmdFlags upgrade.UpgradeValues

var example = fmt.Sprintf(`  # Install GitOps in the %s namespace
  gitops upgrade --profile-version 0.0.15 --app-config-url https://github.com/my-org/my-management-cluster.git`,
	wego.DefaultNamespace)

var Cmd = &cobra.Command{
	Use:           "upgrade",
	Short:         "Upgrade to Weave GitOps Enterprise",
	Example:       example,
	RunE:          upgradeCmdRunE(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.AppConfigURL, "app-config-url", "", "URL of external repository that will hold automation manifests")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfileVersion, "profile-version", "", "Profile version to set the helm release version to")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "base", "main", "The base branch to open the pull request against")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "branch", "tier-upgrade-enterprise", "The branch to create the pull request from")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "commit-message", "Upgrade to WGE", "The commit message")
	Cmd.PersistentFlags().StringArrayVar(&upgradeCmdFlags.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	Cmd.PersistentFlags().BoolVar(&upgradeCmdFlags.DryRun, "dry-run", false, "Output the generated profile without creating a pull request")

	cobra.CheckErr(Cmd.MarkPersistentFlagRequired("app-config-url"))
	cobra.CheckErr(Cmd.MarkPersistentFlagRequired("profile-version"))
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		namespace, err := cmd.Parent().Flags().GetString("namespace")

		if err != nil {
			return fmt.Errorf("couldn't read namespace flag: %v", err)
		}

		// FIXME: maybe a better way to do this?
		upgradeCmdFlags.Namespace = namespace

		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, log)

		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

		gitClient, gitProvider, err := factory.GetGitClients(ctx, providerClient, services.GitConfigParams{
			URL:       upgradeCmdFlags.AppConfigURL,
			Namespace: upgradeCmdFlags.Namespace,
			DryRun:    upgradeCmdFlags.DryRun,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return upgrade.Upgrade(
			ctx,
			gitClient,
			gitProvider,
			upgradeCmdFlags,
			log,
			os.Stdout,
		)
	}
}
