package add

import (
	"bytetrade.io/web3os/minio-operator/cmd/add/driver"
	"bytetrade.io/web3os/minio-operator/cmd/add/node"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [node | driver]",
		Short: "add MinIO node or driver",
		Long:  `add a node or new driver volumes to MinIO cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmd.AddCommand(node.NewCommand())
	cmd.AddCommand(driver.NewCommand())

	return cmd
}
