// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/wuhan005/gadget"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/languages"
)

type KubernetesTask struct {
	config    *rest.Config
	k8sClient *kubernetes.Clientset
	runner    *languages.Runner

	language string
	template *db.Tpl
	code     []byte
}

type NewKubernetesTaskOptions struct {
	Language string
	Template *db.Tpl
	Code     []byte
}

func NewKubernetesTask(_ context.Context, options NewKubernetesTaskOptions) (*KubernetesTask, error) {
	language := options.Language
	template := options.Template
	code := options.Code

	// Set the programming language runner.
	runner, ok := languages.Get(language)
	if !ok {
		return nil, errors.Errorf("unexpected language: %q", language)
	}

	caData := []byte(gadget.Base64Decode(config.App.KubernetesCAData))
	certData := []byte(gadget.Base64Decode(config.App.KubernetesCertData))
	keyData := []byte(gadget.Base64Decode(config.App.KubernetesKeyData))
	bearerToken := config.App.KubernetesBearerToken

	restConfig := &rest.Config{
		Host: config.App.KubernetesServiceHost,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   caData,
			CertData: certData,
			KeyData:  keyData,
		},
		BearerToken: bearerToken,
	}
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}

	return &KubernetesTask{
		config:    restConfig,
		k8sClient: k8sClient,
		runner:    runner,

		language: language,
		template: template,
		code:     code,
	}, nil
}

func (t *KubernetesTask) Run(ctx context.Context) (*ExecOutput, error) {
	namespace := config.App.KubernetesNamespace
	name := "elaina-" + uuid.NewV4().String()
	falseVal := false

	pod, err := t.k8sClient.CoreV1().Pods(namespace).Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"elaina_language": t.language,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           t.runner.Image,
					ImagePullPolicy: v1.PullIfNotPresent,
					SecurityContext: &v1.SecurityContext{
						Privileged:               &falseVal,
						AllowPrivilegeEscalation: &falseVal,
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewQuantity(t.template.MaxCPUs, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(t.template.MaxMemory*1024*1024, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("1m"),
							v1.ResourceMemory: *resource.NewQuantity(5*1024*1024, resource.DecimalSI),
						},
					},
					Stdin:     true,
					StdinOnce: true,
					TTY:       false,
					Env:       nil, // TODO
				},
			},
			AutomountServiceAccountToken: &falseVal,
			EnableServiceLinks:           &falseVal,
			RestartPolicy:                v1.RestartPolicyNever,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "create pod")
	}

	defer func() {
		zeroVal := int64(0)

		if err := t.k8sClient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &zeroVal, // Delete
		}); err != nil {
			logrus.WithError(err).Error("Failed to delete pod")
		}
	}()

	// Wait the pod started.
	if err := waitForPodRunning(ctx, t.k8sClient, pod.Namespace, pod.Name, 60*time.Second); err != nil {
		return nil, errors.Wrap(err, "wait for pod running")
	}

	output, err := t.exec(ctx, pod.Namespace, pod.Name, name, string(t.code))
	if err != nil {
		return nil, errors.Wrap(err, "exec")
	}
	return output, nil
}

func (t *KubernetesTask) exec(ctx context.Context, namespace, podName, containerName string, code string) (*ExecOutput, error) {
	req := t.k8sClient.CoreV1().RESTClient().Post().Resource("pods").
		Namespace(namespace).Name(podName).SubResource("attach").
		VersionedParams(&v1.PodAttachOptions{
			Container: containerName,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(t.config, http.MethodPost, req.URL())
	if err != nil {
		return nil, errors.Wrap(err, "new executor")
	}

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
	stdin, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(err, "marshal input")
	}

	var stdout, stderr bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.template.Timeout)*time.Second)
	defer cancel()

	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  bytes.NewBuffer(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	}); err != nil {
		return nil, errors.Wrap(err, "stream")
	}

	var execOutput ExecOutput
	if err := json.NewDecoder(&stdout).Decode(&execOutput); err != nil {
		return nil, errors.Wrap(err, "unmarshal output")
	}
	return &execOutput, nil
}

// waitForPodRunning polls up to timeout seconds for pod to enter running state.
// Returns an error if the pod never enters the running state.
func waitForPodRunning(ctx context.Context, client kubernetes.Interface, namespace, podName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, timeout, true, isPodRunning(client, podName, namespace))
}

// isPodRunning returns a condition function that indicates whether the given pod is currently running.
func isPodRunning(client kubernetes.Interface, podName, namespace string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
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
