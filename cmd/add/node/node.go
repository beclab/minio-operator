package node

import (
	"os"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "add MinIO node",
		Long:  `add a new node to MinIO cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(); err != nil {
				klog.Errorf("failed to run MinIO add node: %+v", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&base.ETCD_CAFILE, "cafile", "", base.ETCD_CAFILE, "etcd ca file")
	cmd.Flags().StringVarP(&base.ETCD_CERTFILE, "certfile", "", base.ETCD_CERTFILE, "etcd cert file")
	cmd.Flags().StringVarP(&base.ETCD_KEYFILE, "keyfile", "", base.ETCD_KEYFILE, "etcd key file")
	cmd.Flags().StringVarP(&base.ETCD_SERVER, "server", "", base.ETCD_SERVER, "etcd server address")
	cmd.Flags().StringVarP(&base.MINIO_NODE_VOLUME, "volume", "", base.MINIO_NODE_VOLUME, "data volume to be added, ex. /data/vol{1...4}")
	cmd.Flags().StringVarP(&MINIO_NODE_ADDRESS, "address", "", MINIO_NODE_ADDRESS, "minio node ip address")

	return cmd
}

func Run() error {
	cmd := NewNode()

	return cmd.Exec()
}
