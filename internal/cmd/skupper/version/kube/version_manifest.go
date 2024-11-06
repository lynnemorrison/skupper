package kube

import (
	"github.com/skupperproject/skupper/pkg/utils/configs"

	"github.com/skupperproject/skupper/internal/cmd/skupper/common/utils"

	"github.com/skupperproject/skupper/internal/kube/client"
	skupperv2alpha1 "github.com/skupperproject/skupper/pkg/generated/client/clientset/versioned/typed/skupper/v2alpha1"
	"github.com/spf13/cobra"
)

type CmdVersionManifest struct {
	client    skupperv2alpha1.SkupperV2alpha1Interface
	CobraCmd  *cobra.Command
	namespace string
	name      string
	output    string
}

func NewCmdVersionManifest() *CmdVersionManifest {
	return &CmdVersionManifest{}
}

func (cmd *CmdVersionManifest) NewClient(cobraCommand *cobra.Command, args []string) {
	cli, err := client.NewClient(cobraCommand.Flag("namespace").Value.String(), cobraCommand.Flag("context").Value.String(), cobraCommand.Flag("kubeconfig").Value.String())
	utils.HandleError(err)

	cmd.client = cli.GetSkupperClient().SkupperV2alpha1()
	cmd.namespace = cli.Namespace
}

func (cmd *CmdVersionManifest) Run() error {
	manifestManager := configs.ManifestManager{EnableSHA: true}
	return manifestManager.CreateFile(manifestManager.GetDefaultManifestWithEnv())
}

func (cmd *CmdVersionManifest) ValidateInput(args []string) []error { return nil }
func (cmd *CmdVersionManifest) InputToOptions()                     {}
func (cmd *CmdVersionManifest) WaitUntil() error                    { return nil }
