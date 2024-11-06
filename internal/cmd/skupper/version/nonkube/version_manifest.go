package nonkube

import (
	"fmt"

	"github.com/spf13/cobra"
)

type CmdVersionManifest struct {
	CobraCmd  *cobra.Command
	Namespace string
	siteName  string
}

func NewCmdVersionManifest() *CmdVersionManifest {
	return &CmdVersionManifest{}
}

func (cmd *CmdVersionManifest) NewClient(cobraCommand *cobra.Command, args []string) {
	//TODO
}

func (cmd *CmdVersionManifest) ValidateInput(args []string) []error { return nil }
func (cmd *CmdVersionManifest) InputToOptions()                     {}
func (cmd *CmdVersionManifest) Run() error {
	return fmt.Errorf("command not supported by the selected platform")
}
func (cmd *CmdVersionManifest) WaitUntil() error { return nil }
