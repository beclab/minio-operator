package main

import (
	"bytetrade.io/web3os/minio-operator/cmd/add"
	"bytetrade.io/web3os/minio-operator/cmd/init_cmd"
	"bytetrade.io/web3os/minio-operator/cmd/server"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "minio-operator",
		Short: "MinIO Operator",
		Long:  `It is a MinIO Cluster an auto-management service on linux`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	rootCmd.AddCommand(init_cmd.NewCommand())
	rootCmd.AddCommand(add.NewCommand())
	rootCmd.AddCommand(server.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}
