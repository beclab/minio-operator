package init_cmd

import (
	"os"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init MinIO node",
		Long:  `init MinIO node with volumes, volumes should be the path to the data store,like '/data/vol{1...4}'`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(); err != nil {
				klog.Errorf("failed to run MinIO init: %+v", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&base.ETCD_CAFILE, "cafile", "", base.ETCD_CAFILE, "etcd ca file")
	cmd.Flags().StringVarP(&base.ETCD_CERTFILE, "certfile", "", base.ETCD_CERTFILE, "etcd cert file")
	cmd.Flags().StringVarP(&base.ETCD_KEYFILE, "keyfile", "", base.ETCD_KEYFILE, "etcd key file")
	cmd.Flags().StringVarP(&base.ETCD_SERVER, "server", "", base.ETCD_SERVER, "etcd server address")
	cmd.Flags().StringVarP(&base.MINIO_NODE_VOLUME, "volume", "", base.MINIO_NODE_VOLUME, "minio data volume, ex. /data/vol{1...4}")
	cmd.Flags().StringVarP(&MINIO_NODE_ADDRESS, "address", "", MINIO_NODE_ADDRESS, "minio node ip address")
	cmd.Flags().StringVarP(&ROOT_PASSWORD, "password", "", ROOT_PASSWORD, "minio root password")

	return cmd
}

func Run() error {
	cmd := NewInit()

	return cmd.Exec()
}
