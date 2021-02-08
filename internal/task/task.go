package task

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/satori/go.uuid"
)

type Task struct {
	ctx context.Context

	uuid         string
	runner       *runner
	sourceVolume string
}

type Output struct {
	ExitCode int64
	Body     []byte
}

func NewTask(language string, code []byte) (*Task, error) {
	// Check the language.
	var runner *runner
	for _, r := range langRunners {
		if r.Name == language {
			runner = &r
			break
		}
	}
	if runner == nil {
		return nil, errors.Errorf("unexpected error: %v", language)
	}

	id := uuid.NewV4().String()

	// Make runner folder.
	volumePath := path.Join(GetAbsTmpVolume(), id)
	err := os.MkdirAll(volumePath, 0755)
	if err != nil {
		return nil, err
	}

	// Write the code file.
	err = ioutil.WriteFile(path.Join(volumePath, "code"+runner.Ext), code, 0755)
	if err != nil {
		return nil, err
	}

	return &Task{
		ctx:          context.Background(),
		uuid:         id,
		runner:       runner,
		sourceVolume: volumePath,
	}, nil
}

// Run runs a task.
func (t *Task) Run() (*Output, error) {
	client, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}

	resp, err := client.ContainerCreate(t.ctx, &container.Config{
		Image: t.runner.Image,
		Cmd:   t.runner.Cmd,
		Tty:   false,
	}, &container.HostConfig{
		NetworkMode: "none",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: t.sourceVolume,
				Target: "/runner",
			},
		},
		// TODO: Resources
	}, nil, nil, t.uuid)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := client.ContainerStop(t.ctx, resp.ID, nil); err != nil {
			log.Error("Failed to stop container: %v", err)
		}

		if err := client.ContainerRemove(t.ctx, resp.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
			log.Error("Failed to remove container: %v", err)
		}

		err = os.RemoveAll(t.sourceVolume)
		if err != nil {
			log.Error("Failed to remove volume folder: %v", err)
		}
	}()

	if err := client.ContainerStart(t.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	okBody, errChan := client.ContainerWait(t.ctx, resp.ID, "")

	timeout := time.NewTimer(50 * time.Second)
	var waitBody container.ContainerWaitOKBody
	var errExec error
	select {
	case waitBody = <-okBody:
		break
	case errC := <-errChan:
		errExec = errC
	case <-timeout.C:
		errExec = errors.New("execute timeout")
	}
	if errExec != nil {
		return nil, errExec
	}

	// Get the output.
	logs, err := client.ContainerLogs(t.ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, err
	}

	output, err := ioutil.ReadAll(logs)
	if err != nil {
		return nil, err
	}

	return &Output{
		ExitCode: waitBody.StatusCode,
		Body:     output,
	}, nil
}
