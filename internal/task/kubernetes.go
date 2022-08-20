// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
)

type KubernetesTask struct {
	config    *rest.Config
	k8sClient *kubernetes.Clientset
	runner    *runner

	language string
	template *db.Tpl
	code     []byte
}

type NewKubernetesTaskOptions struct {
	Language string
	Template *db.Tpl
	Code     []byte
}

func NewKubernetesTask(ctx context.Context, options NewKubernetesTaskOptions) (*KubernetesTask, error) {
	language := options.Language
	template := options.Template
	code := options.Code

	// Set the programming language runner.
	var runner *runner
	for _, r := range langRunners {
		if r.Name == language {
			runner = &r
			break
		}
	}
	if runner == nil {
		return nil, errors.Errorf("unexpected language: %q", language)
	}

	config := &rest.Config{
		Host: os.Getenv("KUBERNETES_SERVICE_HOST"),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}

	return &KubernetesTask{
		config:    config,
		k8sClient: k8sClient,
		runner:    runner,
		language:  language,
		template:  template,
		code:      code,
	}, nil
}

func (t *KubernetesTask) Run(ctx context.Context) ([]*CommandOutput, error) {
	namespace := os.Getenv("KUBERNETES_NAMESPACE")
	name := "elaina-" + uuid.NewV4().String()
	falseVal := false

	pod, err := t.k8sClient.CoreV1().Pods(namespace).Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"elaina_language": t.language,
				"elaina_template": t.template.Name,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           t.runner.Image,
					ImagePullPolicy: v1.PullIfNotPresent,
					SecurityContext: &v1.SecurityContext{
						AllowPrivilegeEscalation: &falseVal,
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewQuantity(t.template.MaxCPUs, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(t.template.MaxMemory*1024*1024, resource.DecimalSI),
						},
					},
				},
			},
			AutomountServiceAccountToken: &falseVal,
			EnableServiceLinks:           &falseVal,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "create pod")
	}

	defer func() {
		if err := t.k8sClient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
			log.Error("Failed to delete pod: %v", err)
		}
	}()

	// Wait the pod started.
	if err := waitForPodRunning(ctx, t.k8sClient, pod.Namespace, pod.Name, time.Duration(t.template.Timeout)*time.Second); err != nil {
		return nil, errors.Wrap(err, "wait for pod running")
	}

	// Write the code to the container.
	filePath := filepath.Join("/code/" + "code" + t.runner.Ext)
	cmd := []string{"sh", "-c", "echo '" + string(t.code) + "' > " + filePath}
	_, err = t.exec(ctx, name, namespace, cmd)
	if err != nil {
		return nil, errors.Wrap(err, "exec: write code file")
	}

	output := make([]*CommandOutput, 0, 2)
	if t.runner.BuildCmd != "" {
		buildOutput, err := t.exec(ctx, name, namespace, []string{"sh", "-c", t.runner.BuildCmd})
		if err != nil {
			return nil, errors.Wrap(err, "exec: build")
		}
		output = append(output, buildOutput)
	}

	runOutput, err := t.exec(ctx, name, namespace, []string{"sh", "-c", t.runner.RunCmd})
	if err != nil {
		return nil, errors.Wrap(err, "exec: run")
	}
	output = append(output, runOutput)

	return output, nil
}

func (t *KubernetesTask) exec(ctx context.Context, name, namespace string, cmd []string) (*CommandOutput, error) {
	req := t.k8sClient.CoreV1().RESTClient().Post().Resource("pods").Name(name).Namespace(namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
		Command: cmd,
	}
	req.VersionedParams(option, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(t.config, http.MethodPost, req.URL())
	if err != nil {
		return nil, errors.Wrap(err, "new executor")
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "stream")
	}

	errResp := stderr.Bytes()
	if len(errResp) != 0 {
		return &CommandOutput{
			ExitCode: 1,
			Body:     errResp,
		}, nil
	}

	return &CommandOutput{
		ExitCode: 0,
		Body:     stdout.Bytes(),
	}, nil
}

// waitForPodRunning polls up to timeout seconds for pod to enter running state.
// Returns an error if the pod never enters the running state.
func waitForPodRunning(ctx context.Context, client kubernetes.Interface, namespace, podName string, timeout time.Duration) error {
	return wait.PollImmediate(100*time.Millisecond, timeout, isPodRunning(ctx, client, podName, namespace))
}

// isPodRunning returns a condition function that indicates whether the given pod is currently running.
func isPodRunning(ctx context.Context, client kubernetes.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, errors.New("pod is finished")
		}
		return false, nil
	}
}
