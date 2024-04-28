# minio-operator
[![](https://github.com/Above-Os/minio-operator/actions/workflows/build.yaml/badge.svg?branch=main)](https://github.com/Above-Os/minio-operator/actions/workflows/build.yaml)

## Description
MinIO Operator provides a command line tool for the MinIO cluster maintenance. The Operator offers the `init`, `add node`, `add driver` commands to create a cluster and scale it to fulfill the different scenarios.

## How to use
1. Install ETCD as the operator metadata storage
   [Install ETCD](https://etcd.io/docs/v3.5/install/)

2. Install MinIO
```sh

MINIO_VERSION=<version to install>
curl -kLo minio https://dl.min.io/server/minio/release/linux-amd64/archive/minio.${MINIO_VERSION}
sudo install -m 755 minio /usr/local/bin/minio

```

3. Install MinIO Operator
```sh

MINIO_OPERATOR_VERSION="v0.0.1"
curl -k -sfLO https://github.com/Above-Os/minio-operator/releases/download/${MINIO_OPERATOR_VERSION}/minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz
tar zxf minio-operator-${MINIO_OPERATOR_VERSION}-linux-amd64.tar.gz
sudo install -m 755 minio-operator /usr/local/bin/operator

```

4. Init cluster
```sh

local_ip=<YOUR MACHINE IP>
MINIO_VOLUMES="/path/to/minio/driver{1...4}"

sudo groupadd -r minio
sudo useradd -M -r -g minio minio
sudo chown minio:minio /path/to/minio/driver{1..4}

sudo minio-operator init --address $local_ip \
   --cafile /etc/ssl/etcd/ssl/ca.pem \
   --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem \
   --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem \
   --volume $MINIO_VOLUMES

```
According to MinIO [official document](https://min.io/docs/minio/linux/operations/install-deploy-manage/deploy-minio-single-node-multi-drive.html#minio-snmd), MinIO strongly recommends provisioning XFS formatted drives for storage. And every single volume path should be mounted on a separate driver.

5. After cluster init, a SNMD mode MinIO cluster will be running at the current machine

### Add Driver

```sh
VOLUMES="/path/to/minio/driver{5...6}"

sudo chown minio:minio /path/to/minio/driver{5..6}

sudo minio-operator add driver --cafile /etc/ssl/etcd/ssl/ca.pem \
   --certfile /etc/ssl/etcd/ssl/node-$HOSTNAME.pem \
   --keyfile /etc/ssl/etcd/ssl/node-$HOSTNAME-key.pem \
   --volume $VOLUMES

```

### Add Node
```sh

VOLUMES="/path/to/minio/driver{1...4}"
local_ip=<YOUR MACHINE IP>
MASTER_NODE=<MASTER NODE IP>
ETCD_SERVER="${MASTER_NODE}:2379"
MASTER_NODE_HOSTNAME=<MASTER NODE HOSTNAME>

sudo chown minio:minio /path/to/minio/driver{1..4}

sudo minio-operator add node --server ${ETCD_SERVER} \
   --address $local_ip \
   --cafile /etc/ssl/etcd/ssl/ca.pem \
   --certfile /etc/ssl/etcd/ssl/node-$MASTER_NODE_HOSTNAME.pem \
   --keyfile /etc/ssl/etcd/ssl/node-$MASTER_NODE_HOSTNAME-key.pem \
   --volume $VOLUMES

```

## How to build

```sh
make build
```

or 

```sh
make build-linux
```
for Linux version