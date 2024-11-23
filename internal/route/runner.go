// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"time"

	"github.com/flamego/template"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/runtime"
)

type RunnerHandler struct{}

func NewRunnerHandler() *RunnerHandler {
	return &RunnerHandler{}
}

func (h *RunnerHandler) Runner(ctx context.Context) error {
	sandboxUID := ctx.Param("uid")
	sandbox, err := db.Sandboxes.GetByUID(ctx.Request().Context(), sandboxUID)
	if err != nil {
		ctx.Redirect("/")
		return nil
	}

	ctx.Map(sandbox)
	return nil
}

func (h *RunnerHandler) View(ctx context.Context, sandbox *db.Sandbox, t template.Template, data template.Data) {
	languages := sandbox.Template.Language
	selectedLanguage := ctx.Query("l")
	if !lo.Contains(languages, selectedLanguage) {
		selectedLanguage = languages[0]
	}

	_ = ctx.Request().ParseForm()
	code := ctx.Request().PostForm.Get("c")
	if code == "" {
		code = sandbox.Placeholder
	}

	data["Sandbox"] = sandbox
	data["Language"] = selectedLanguage
	data["Languages"] = languages
	data["Code"] = code

	t.HTML(http.StatusOK, "sandbox")
}

func (h *RunnerHandler) Execute(ctx context.Context, sandbox *db.Sandbox) error {
	if err := ctx.Request().ParseForm(); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to parse form")
		return ctx.Error(http.StatusBadRequest, "Failed to parse form: %v", err)
	}

	templateLanguages := sandbox.Template.Language
	selectedLanguage := ctx.Request().PostForm.Get("lang")
	if !lo.Contains(templateLanguages, selectedLanguage) {
		selectedLanguage = templateLanguages[0]
	}
	code := ctx.Request().PostForm.Get("code")

	// TODO: Rate limit

	startAt := time.Now().UnixNano()

	var r runtime.ExecRuntime
	var err error

	switch config.App.RuntimeMode {
	case "kubernetes", "k8s":
		r, err = runtime.NewKubernetesTask(ctx.Request().Context(), runtime.NewKubernetesTaskOptions{
			Language: selectedLanguage,
			Template: sandbox.Template,
			Code:     []byte(code),
		})

	case "docker", "":
		r, err = runtime.NewDockerTask(ctx.Request().Context(), runtime.NewDockerTaskOptions{
			Language: selectedLanguage,
			Template: sandbox.Template,
			Code:     []byte(code),
		})

	default:
		return ctx.Error(http.StatusInternalServerError, "unexpected runtime mode: %q", config.App.RuntimeMode)
	}

	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create task")
		return ctx.Error(http.StatusInternalServerError, "Failed to create task: %v", err)
	}

	output, err := r.Run(ctx.Request().Context())
	if err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to run task")
		return ctx.Error(http.StatusInternalServerError, "Failed to run task: %v", err)
	}

	endAt := time.Now().UnixNano()

	return ctx.Success(map[string]interface{}{
		"result":   output,
		"start_at": startAt,
		"end_at":   endAt,
	})
}
