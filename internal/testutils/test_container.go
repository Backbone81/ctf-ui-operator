package testutils

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/controller/minio"
)

const (
	MinioUser     = "minio"
	MinioPassword = "minio123"
)

func NewCTFdTestContainer(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: ctfd.Image,
			ExposedPorts: []string{
				"8000",
			},
			WaitingFor: wait.ForLog("Listening at: http://0.0.0.0:8000"),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func NewMinioTestContainer(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: minio.Image,
			Cmd: []string{
				"server",
				"--address=:9000",
				"--console-address=:9001",
				"/mnt/data",
			},
			ExposedPorts: []string{
				"9000",
			},
			Env: map[string]string{
				"MINIO_ROOT_USER":     MinioUser,
				"MINIO_ROOT_PASSWORD": MinioPassword,
			},
			WaitingFor: wait.ForLog("Docs: https://docs.min.io"),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}
