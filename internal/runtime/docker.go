package runtime

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/languages"
)

type DockerTask struct {
	uuid        string
	containerID string

	runner   *languages.Runner
	template *db.Tpl
	code     []byte

	dockerClient *client.Client
}

type NewDockerTaskOptions struct {
	Language string
	Template *db.Tpl
	Code     []byte
}

// NewDockerTask creates a new task based on the given code and ready for execution.
func NewDockerTask(ctx context.Context, options NewDockerTaskOptions) (*DockerTask, error) {
	uid := uuid.NewV4().String()
	language := options.Language
	template := options.Template
	code := options.Code

	// Set the programming language runner.
	runner, ok := languages.Get(language)
	if !ok {
		return nil, errors.Errorf("unexpected language: %q", language)
	}

	// Create a new docker client.
	dockerClient, err := client.NewClientWithOpts()
	if err != nil {
		return nil, errors.Wrap(err, "new docker client")
	}
	dockerClient.NegotiateAPIVersion(ctx)

	return &DockerTask{
		uuid:         uid,
		runner:       runner,
		template:     template,
		code:         code,
		dockerClient: dockerClient,
	}, nil
}

// Run runs a task.
func (t *DockerTask) Run(ctx context.Context) (*ExecOutput, error) {
	createContainerResp, err := t.dockerClient.ContainerCreate(ctx,
		&container.Config{
			Hostname:        "elaina-runtime",
			AttachStdin:     true,
			AttachStdout:    true,
			AttachStderr:    true,
			Tty:             false,
			OpenStdin:       true,
			StdinOnce:       true,
			Image:           t.runner.Image,
			NetworkDisabled: !t.template.InternetAccess,
			Env:             nil, // TODO
		},
		&container.HostConfig{
			Resources: container.Resources{
				NanoCPUs: t.template.MaxCPUs * 1000000000,    // 0.000000001 * CPU of cpu
				Memory:   t.template.MaxMemory * 1024 * 1024, // Minimum memory limit allowed is 6MB.
			},
		}, nil, nil, t.uuid)
	if err != nil {
		return nil, errors.Wrap(err, "create container")
	}

	t.containerID = createContainerResp.ID

	// Clean containers and folder after executed.
	defer t.clean(ctx)

	if err := t.dockerClient.ContainerStart(ctx, t.containerID, container.StartOptions{}); err != nil {
		return nil, errors.Wrap(err, "start container")
	}

	// Execute code.
	output, err := t.exec(ctx, string(t.code))
	if err != nil {
		return nil, errors.Wrap(err, "exec")
	}
	return output, nil
}

func (t *DockerTask) exec(ctx context.Context, code string) (*ExecOutput, error) {
	attach, err := t.dockerClient.ContainerAttach(ctx, t.containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "attach")
	}
	defer func() { attach.Close() }()

	buildCommands := t.runner.BuildCommands
	if buildCommands == nil {
		buildCommands = []string{}
	}

	input := ExecInput{
		RunInstructions: ExecInputRunInstructions{
			BuildCommands: buildCommands,
			RunCommand:    t.runner.RunCommand,
		},
		Files: []*ExecInputFile{
			{
				Name:    t.runner.FileName,
				Content: code,
			},
		},
		Stdin: nil,
	}

	if err := json.NewEncoder(attach.Conn).Encode(input); err != nil {
		return nil, errors.Wrap(err, "write")
	}

	// Send an `EOF` to writer.
	if err := attach.CloseWrite(); err != nil {
		return nil, errors.Wrap(err, "close write")
	}

	output, err := io.ReadAll(attach.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "read output")
	}
	output = parseDockerLog(output)

	var execOutput ExecOutput
	if err := json.Unmarshal(output, &execOutput); err != nil {
		return nil, errors.Wrap(err, "unmarshal output")
	}
	return &execOutput, nil
}

func (t *DockerTask) clean(ctx context.Context) {
	if err := t.dockerClient.ContainerRemove(ctx, t.containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to remove container")
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
