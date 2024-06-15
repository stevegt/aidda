package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Function to create Docker client
func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// Function to execute action in Docker container
func executeActionInContainer(image string, action Action) (string, error) {
	cli, err := createDockerClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	pwd := os.Getenv("PWD")
	hostConfig := &container.HostConfig{
		Binds: []string{pwd + ":/mnt"},
	}

	ctx := context.Background()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Tty:   false,
	}, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	defer cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	logs, err := ioutil.ReadAll(out)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}
