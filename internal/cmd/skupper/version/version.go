package version

import (
	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/internal/cmd/skupper/common"
	"github.com/skupperproject/skupper/internal/cmd/skupper/version/kube"
	"github.com/skupperproject/skupper/internal/cmd/skupper/version/nonkube"
	"github.com/skupperproject/skupper/pkg/config"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display versions of Skupper components.",
		Long:  "Report the version of the Skupper CLI and services",
		Example: `skupper version status
skupper version manifest`,
	}

	cmd.AddCommand(CmdVersionFactory(config.GetPlatform()))
	cmd.AddCommand(CmdVersionManifestFactory(config.GetPlatform()))

	return cmd
}

func CmdVersionFactory(configuredPlatform types.Platform) *cobra.Command {
	kubeCommand := kube.NewCmdVersion()
	nonKubeCommand := nonkube.NewCmdVersion()

	cmdVersionDesc := common.SkupperCmdDescription{
		Use:   "status",
		Short: "Report software versions",
		Long:  "Report the version of the Skupper CLI and services",
	}

	cmd := common.ConfigureCobraCommand(configuredPlatform, cmdVersionDesc, kubeCommand, nonKubeCommand)

	cmdFlags := common.CommandVersionFlags{}
	cmd.Flags().StringVarP(&cmdFlags.Output, common.FlagNameOutput, "o", "", common.FlagDescOutput)

	kubeCommand.CobraCmd = cmd
	kubeCommand.Flags = &cmdFlags
	nonKubeCommand.CobraCmd = cmd
	nonKubeCommand.Flags = &cmdFlags

	return cmd
}

func CmdVersionManifestFactory(configuredPlatform types.Platform) *cobra.Command {
	kubeCommand := kube.NewCmdVersionManifest()
	nonKubeCommand := nonkube.NewCmdVersionManifest()

	cmdVersionDesc := common.SkupperCmdDescription{
		Use:   "manifest",
		Short: "Generate file of software versions",
		Long:  "Generate a JSON file containing the version of the Skupper images by default and the value of the environment variables in the current directory.",
	}

	cmd := common.ConfigureCobraCommand(configuredPlatform, cmdVersionDesc, kubeCommand, nonKubeCommand)

	kubeCommand.CobraCmd = cmd
	nonKubeCommand.CobraCmd = cmd

	return cmd
}
