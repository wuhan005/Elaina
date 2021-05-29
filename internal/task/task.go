package task

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
)

var hostVolumePath = path.Join(os.Getenv("APP_CONTAINER_PATH"), "volume")

type Task struct {
	ctx context.Context

	uuid     string
	runner   *runner
	template *db.Tpl

	dockerClient *client.Client
	containerID  string

	sourceVolumePath string // Folder in Host: /home/<your_user>/elaina/volume/<uuid>/
	elainaVolumePath string // Folder in Elaina: /elaina/volume/<uuid>/
	fileName         string
}

type Output struct {
	Error bool   `json:"error"`
	Body  []byte `json:"body"`
}

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

	sourceVolumePath := path.Join(hostVolumePath, uid)
	// Make runner folder.
	elainaVolumePath := path.Join("/elaina/volume", uid)
	err = os.MkdirAll(elainaVolumePath, 0755)
	if err != nil {
		return nil, err
	}

	// Make the `runner` folder and create code file, `code.<ext>`.
	runnerPath := path.Join(elainaVolumePath, "runner")
	err = os.MkdirAll(runnerPath, 0755)
	if err != nil {
		return nil, err
	}
	fileName := "code" + runner.Ext
	filePath := path.Join(runnerPath, fileName)
	err = ioutil.WriteFile(filePath, code, 0755)
	if err != nil {
		return nil, err
	}

	return &Task{
		ctx: ctx,

		uuid:     uid,
		runner:   runner,
		template: template,

		dockerClient: dockerClient,

		sourceVolumePath: sourceVolumePath,
		elainaVolumePath: elainaVolumePath,

		fileName: fileName,
	}, nil
}

// Run runs a task.
func (t *Task) Run() ([]*Output, error) {
	output := make([]*Output, 0, 2)

	var networkMode container.NetworkMode
	if t.template.InternetAccess {
		networkMode = "bridge"
	} else {
		networkMode = "none"
	}

	createContainerResp, err := t.dockerClient.ContainerCreate(t.ctx,
		&container.Config{
			Image: t.runner.Image,
			Tty:   true,
		},
		&container.HostConfig{
			NetworkMode: networkMode,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: t.sourceVolumePath,
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

	setupOutput, err := t.setupEnvironment()
	if err != nil {
		return output, err
	}
	output = append(output, setupOutput)
	if setupOutput.Error {
		return output, nil
	}

	// Execute code.
	runOutput, err := t.exec(t.runner.RunCmd)
	if err != nil {
		return output, err
	}
	output = append(output, runOutput)

	return output, nil
}

func (t *Task) setupEnvironment() (*Output, error) {
	if len(t.runner.BuildCmd) != 0 {
		return t.exec(t.runner.BuildCmd)
	}
	return &Output{}, nil
}

func (t *Task) exec(cmd string) (*Output, error) {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", path.Join(t.elainaVolumePath, "elaina-daemon.sock"))
			},
		},
	}

	cmdJSON, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	for { // Retry, wait for daemon starts.
		resp, err = client.Post("http://runtime/exec", "", bytes.NewReader(cmdJSON))
		if err != nil {
			if strings.HasSuffix(err.Error(), "connect: no such file or directory") ||
				strings.HasSuffix(err.Error(), "connect: connection refused") {
				continue
			}
			return nil, err
		}
		break
	}

	defer func() {
		_ = resp.Body.Close()
		_, _ = io.Copy(ioutil.Discard, resp.Body)
	}()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var cmdResp struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Error  bool   `json:"error"`
	}
	err = json.Unmarshal(respBody, &cmdResp)
	if err != nil {
		return nil, err
	}

	var body string
	if cmdResp.Stderr != "" {
		body = cmdResp.Stderr
	} else {
		body = cmdResp.Stdout
	}

	return &Output{
		Error: cmdResp.Error,
		Body:  []byte(body),
	}, nil
}

func (t *Task) clean() {
	if err := t.dockerClient.ContainerStop(t.ctx, t.containerID, nil); err != nil {
		log.Error("Failed to stop container: %v", err)
	}

	if err := t.dockerClient.ContainerRemove(t.ctx, t.containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		log.Error("Failed to remove container: %v", err)
	}

	err := os.RemoveAll(path.Join("/elaina/volume", t.uuid))
	if err != nil {
		log.Error("Failed to remove volume folder: %v", err)
	}
}
