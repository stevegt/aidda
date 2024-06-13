package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func runInContainer(testArgs string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	image := "your_image_name"

	// Create a container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"go", "mod", "tidy"},
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	fmt.Println("go mod tidy completed")

	// Copy current directory to the container
	copyToContainer(ctx, cli, resp.ID, ".")

	// Run golint in the container
	execID, err := cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
		Cmd:          []string{"golint"},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		panic(err)
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}
	defer attachResp.Close()

	io.Copy(os.Stdout, attachResp.Reader)

	// Run tests in the container
	execID, err = cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
		Cmd:          []string{"go", "test", "-v", testArgs},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		panic(err)
	}

	attachResp, err = cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}
	defer attachResp.Close()

	io.Copy(os.Stdout, attachResp.Reader)
}

func cleanupContainers(tmpContainerImage string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Stop and remove containers
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "label", Value: "aidda-delete-me"}),
	})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		options := container.StopOptions{}
		err = cli.ContainerStop(ctx, c.ID, options)
		if err != nil {
			panic(err)
		}
		err = cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{})
		if err != nil {
			panic(err)
		}
	}

	// Remove images
	_, err = cli.ImageRemove(ctx, tmpContainerImage, image.RemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		fmt.Println("Error removing image:", err)
	}
}

func tidyAndCommitContainer(containerImage, tmpContainerImage string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create and start the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: containerImage,
		Cmd:   []string{"go", "mod", "tidy"},
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	// Commit the container
	_, err = cli.ContainerCommit(ctx, resp.ID, container.CommitOptions{
		Reference: tmpContainerImage,
	})
	if err != nil {
		panic(err)
	}
}

func runTests(tmpContainerImage, testArgs string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Create a container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: tmpContainerImage,
		Cmd:   []string{"/tmp/aidda", "-Z", testArgs},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: getCurrentDirectory(),
				Target: "/mnt",
			},
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		panic(err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	// Attach to the container
	attachResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		panic(err)
	}
	defer attachResp.Close()

	io.Copy(os.Stdout, attachResp.Reader)

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}
}

func copyToContainer(ctx context.Context, cli *client.Client, containerID, srcPath string) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	filepath.Walk(srcPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(file, srcPath)
		tarHeader := &tar.Header{
			Name: relPath,
			Size: fi.Size(),
			Mode: int64(fi.Mode()),
		}

		if err := tw.WriteHeader(tarHeader); err != nil {
			return err
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	tarReader := bytes.NewReader(buf.Bytes())
	err := cli.CopyToContainer(ctx, containerID, "/", tarReader, types.CopyToContainerOptions{})
	if err != nil {
		panic(err)
	}
}

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}
