package minio

var MinIOService = `
[Unit]
Description=MinIO
Documentation=https://min.io/docs/minio/linux/index.html
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable=/usr/local/bin/minio

[Service]
WorkingDirectory=/usr/local

User=minio
Group=minio
ProtectProc=invisible

EnvironmentFile=-/etc/default/minio
ExecStartPre=/bin/bash -c "if [ -z \"${MINIO_VOLUMES}\" ]; then echo \"Variable MINIO_VOLUMES not set in /etc/default/minio\"; exit 1; fi"
ExecStart=/usr/local/bin/minio server $MINIO_OPTS $MINIO_VOLUMES

# MinIO RELEASE.2023-05-04T21-44-30Z adds support for Type=notify (https://www.freedesktop.org/software/systemd/man/systemd.service.html#Type=)
# This may improve systemctl setups where other services use After=minio.server
# Uncomment the line to enable the functionality
# Type=notify

# Let systemd restart this service always
Restart=always

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

# Specifies the maximum number of threads this process can create
TasksMax=infinity

# Disable timeout logic and wait until process is stopped
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
`

var MinIOEnv = map[string]string{
	"MINIO_ROOT_USER":     DefaultRootUser,
	"MINIO_ROOT_PASSWORD": "",
	"MINIO_VOLUMES":       "/terminus/data/minio/vol1",
	"MINIO_OPTS":          "--console-address :9090",
	"MINIO_SERVER_URL":    "http://minio:9000",
	"CI":                  "true",
	"MINIO_CI_CD":         "true",
}

var MinIOOperatorService = `
[Unit]
Description=MinIO Operator
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable=/usr/local/bin/minio-operator

[Service]
WorkingDirectory=/usr/local

User=root
Group=root
ProtectProc=invisible

EnvironmentFile=-/etc/default/minio-operator
ExecStart=/usr/local/bin/minio-operator server

Restart=always
LimitNOFILE=65536
TasksMax=infinity
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
`
