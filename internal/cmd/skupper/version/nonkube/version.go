package nonkube

import (
	"fmt"

	"github.com/skupperproject/skupper/internal/cmd/skupper/common"
	"github.com/spf13/cobra"
)

type CmdVersion struct {
	CobraCmd  *cobra.Command
	Flags     *common.CommandVersionFlags
	Namespace string
	siteName  string
}

func NewCmdVersion() *CmdVersion {
	return &CmdVersion{}
}

func (cmd *CmdVersion) NewClient(cobraCommand *cobra.Command, args []string) {
	//TODO
}

func (cmd *CmdVersion) ValidateInput(args []string) []error { return nil }
func (cmd *CmdVersion) InputToOptions()                     {}
func (cmd *CmdVersion) Run() error {
	return fmt.Errorf("command not supported by the selected platform")
}
func (cmd *CmdVersion) WaitUntil() error { return nil }
