package kube

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/skupperproject/skupper/pkg/utils/configs"

	"github.com/skupperproject/skupper/internal/cmd/skupper/common"
	"github.com/skupperproject/skupper/internal/cmd/skupper/common/utils"

	"github.com/skupperproject/skupper/internal/kube/client"
	skupperv2alpha1 "github.com/skupperproject/skupper/pkg/generated/client/clientset/versioned/typed/skupper/v2alpha1"
	"github.com/skupperproject/skupper/pkg/utils/validator"
	"github.com/spf13/cobra"
)

type versionOutput struct {
	Component string `json:"component"`
	Version   string `json:"version"`
	Image     string `json:"image"`
	Digest    string `json:"digest"`
}

type CmdVersion struct {
	client    skupperv2alpha1.SkupperV2alpha1Interface
	CobraCmd  *cobra.Command
	Flags     *common.CommandVersionFlags
	namespace string
	name      string
	output    string
}

func NewCmdVersion() *CmdVersion {
	return &CmdVersion{}
}

func (cmd *CmdVersion) NewClient(cobraCommand *cobra.Command, args []string) {
	cli, err := client.NewClient(cobraCommand.Flag("namespace").Value.String(), cobraCommand.Flag("context").Value.String(), cobraCommand.Flag("kubeconfig").Value.String())
	utils.HandleError(err)

	cmd.client = cli.GetSkupperClient().SkupperV2alpha1()
	cmd.namespace = cli.Namespace
}

func (cmd *CmdVersion) ValidateInput(args []string) []error {
	var validationErrors []error
	outputTypeValidator := validator.NewOptionValidator(common.OutputTypes)

	if cmd.Flags != nil && cmd.Flags.Output != "" {
		ok, err := outputTypeValidator.Evaluate(cmd.Flags.Output)
		if !ok {
			validationErrors = append(validationErrors, fmt.Errorf("output type is not valid: %s", err))
		} else {
			cmd.output = cmd.Flags.Output
		}
	}

	return validationErrors
}

func (cmd *CmdVersion) Run() error {
	if cmd.output != "" {
		manifestManager := configs.ManifestManager{EnableSHA: true}
		files := manifestManager.GetConfiguredManifest()
		for _, resource := range files.Images {
			parts := strings.Split(resource.Name, "/")
			last := len(parts)
			if last != 0 {
				nameAndTag := strings.Split(parts[last-1], ":")
				output := versionOutput{
					Component: nameAndTag[0],
					Version:   nameAndTag[1],
					Image:     resource.Name,
					Digest:    resource.SHA,
				}
				encodedOutput, err := utils.Encode(cmd.output, output)
				if err != nil {
					return err
				}
				fmt.Println(encodedOutput)
			}
		}
	} else {
		manifestManager := configs.ManifestManager{EnableSHA: false}
		files := manifestManager.GetConfiguredManifest()
		tw := tabwriter.NewWriter(os.Stdout, 8, 8, 1, '\t', tabwriter.TabIndent)
		_, _ = fmt.Fprintln(tw, fmt.Sprintf("%s\t%s", "COMPONENT", "VERSION"))
		for _, file := range files.Images {
			parts := strings.Split(file.Name, "/")
			last := len(parts)
			if last != 0 {
				nameAndTag := strings.Split(parts[last-1], ":")
				fmt.Fprintln(tw, fmt.Sprintf("%s\t%s", nameAndTag[0], nameAndTag[1]))
			}
		}
		_ = tw.Flush()
	}

	return nil
}

func (cmd *CmdVersion) InputToOptions()  {}
func (cmd *CmdVersion) WaitUntil() error { return nil }
