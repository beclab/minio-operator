package server

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "MinIO Operator Daemon",
		Long:  `MinIO Operator Daemon, a management server for minio cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			server := NewServer()
			if err := server.Exec(); err != nil {
				klog.Errorf("failed to run MinIO init: %+v", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
