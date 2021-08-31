package task

import (
	"context"
	"encoding/binary"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
)

type Task struct {
	ctx context.Context

	uuid     string
	runner   *runner
	template *db.Tpl

	dockerClient *client.Client
	containerID  string

	// sourceAbsVolumePath is the absolute path of the folder in host: <base path>/volume/<uuid>/
	sourceAbsVolumePath string
	fileName            string
}

// commandOutput contains the body and the exit code of the command execution.
type commandOutput struct {
	ExitCode int    `json:"exit_code"`
	Body     []byte `json:"body"`
}

// NewTask creates a new task based on the given code and ready for execution.
func NewTask(language string, template *db.Tpl, code []byte) (*Task, error) {
	uid := uuid.NewV4().String()

	// Set the programming language runner.
	var runner *runner
	for _, r := range langRunners {
		if r.Name == language {
			runner = &r
			break
		}
	}
	if runner == nil {
		return nil, errors.Errorf("unexpected language: %v", language)
	}

	// Create a new docker client.
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	dockerClient.NegotiateAPIVersion(ctx)

	volumeAbsPath, err := filepath.Abs("./volume")
	if err != nil {
		return nil, errors.Wrap(err, "get absolute host volume path")
	}

	sourceAbsVolumePath := path.Join(volumeAbsPath, uid)
	err = os.MkdirAll(sourceAbsVolumePath, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "make source volume path")
	}

	// Make the `runner` folder and create code file, `code.<ext>`.
	fileName := "code" + runner.Ext
	filePath := path.Join(sourceAbsVolumePath, fileName)
	err = os.WriteFile(filePath, code, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "write file")
	}

	return &Task{
		ctx: ctx,

		uuid:     uid,
		runner:   runner,
		template: template,

		dockerClient: dockerClient,

		sourceAbsVolumePath: sourceAbsVolumePath,

		fileName: fileName,
	}, nil
}

// Run runs a task.
func (t *Task) Run() ([]*commandOutput, error) {
	output := make([]*commandOutput, 0, 2) // One for build command, one for run command.

	var networkMode container.NetworkMode
	if t.template.InternetAccess {
		networkMode = "bridge"
	} else {
		networkMode = "none"
	}

	createContainerResp, err := t.dockerClient.ContainerCreate(t.ctx,
		&container.Config{
			Image:        t.runner.Image,
			User:         "elaina",
			WorkingDir:   "/runtime",
			Tty:          false,
			AttachStdout: true,
			AttachStderr: true,
			Env:          nil, // TODO
		},
		&container.HostConfig{
			NetworkMode: networkMode,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: t.sourceAbsVolumePath,
					Target: "/runtime",
				},
			},
			Resources: container.Resources{
				NanoCPUs: t.template.MaxCPUs * 1000000000,    // 0.000000001 * CPU of cpu
				Memory:   t.template.MaxMemory * 1024 * 1024, // Minimum memory limit allowed is 6MB.
			},
		}, nil, nil, t.uuid)
	if err != nil {
		return nil, err
	}
	t.containerID = createContainerResp.ID

	// Clean containers and folder after executed.
	defer t.clean()

	if err := t.dockerClient.ContainerStart(t.ctx, t.containerID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	// Build environment.
	buildOutput, err := t.exec(t.ctx, t.runner.BuildCmd)
	if err != nil {
		return output, errors.Wrap(err, "build")
	}
	output = append(output, buildOutput)

	if buildOutput.ExitCode != 0 {
		return output, nil
	}

	// Execute code.
	runOutput, err := t.exec(t.ctx, t.runner.RunCmd)
	if err != nil {
		return output, errors.Wrap(err, "exec")
	}
	output = append(output, runOutput)

	return output, nil
}

func (t *Task) exec(ctx context.Context, cmd string) (*commandOutput, error) {
	if cmd == "" {
		return &commandOutput{}, nil
	}

	execResp, err := t.dockerClient.ContainerExecCreate(ctx, t.containerID, types.ExecConfig{
		User:         "elaina",
		Tty:          false,
		AttachStderr: true,
		AttachStdout: true,
		Env:          nil, // TODO add environment variables
		WorkingDir:   "/runtime",
		Cmd:          []string{"sh", "-c", cmd},
	})
	if err != nil {
		return nil, errors.Wrap(err, "exec create")
	}

	attachResp, err := t.dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "exec attach")
	}

	defer func() { attachResp.Close() }()

	body, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "read response")
	}

	// Check out the execution status.
	inspectResp, err := t.dockerClient.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, errors.Wrap(err, "exec inspect")
	}

	return &commandOutput{
		ExitCode: inspectResp.ExitCode,
		Body:     parseDockerLog(body),
	}, nil
}

func (t *Task) clean() {
	if err := t.dockerClient.ContainerKill(t.ctx, t.containerID, "9"); err != nil {
		log.Error("Failed to kill container: %v", err)
	}

	if err := t.dockerClient.ContainerRemove(t.ctx, t.containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		log.Error("Failed to remove container: %v", err)
	}

	err := os.RemoveAll(t.sourceAbsVolumePath)
	if err != nil {
		log.Error("Failed to remove volume folder: %v", err)
	}
}

// parseDockerLog parse the header of the docker logs.
// More information at: https://github.com/moby/moby/issues/7375#issuecomment-51462963
//
// header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
func parseDockerLog(logs []byte) []byte {
	output := make([]byte, 0, len(logs))

	for i := 0; i < len(logs); {
		sizeBinary := logs[i+4 : i+8]
		i += 8

		size := int(binary.BigEndian.Uint32(sizeBinary))
		data := logs[i : i+size]
		output = append(output, data...)
		i += size
	}

	return output
}
