// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"time"

	"github.com/flamego/template"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/runtime"
)

type RunnerHandler struct{}

type ExecutionResult struct {
	Result  *runtime.ExecOutput `json:"result"`
	StartAt int64               `json:"start_at"`
	EndAt   int64               `json:"end_at"`
} // @name ExecutionResult

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

// selectLanguage returns language if the template supports it, otherwise the first supported language.
func selectLanguage(languages []string, language string) string {
	if lo.Contains(languages, language) {
		return language
	}
	return languages[0]
}

func (h *RunnerHandler) View(ctx context.Context, sandbox *db.Sandbox, t template.Template, data template.Data) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.Error(http.StatusBadRequest, "Failed to parse form")
	}
	code := ctx.Request().PostForm.Get("c")
	if code == "" {
		code = sandbox.Placeholder
	}
	if int64(len(code)) > config.App.WASMLimits.MaxSourceBytes {
		return ctx.Error(http.StatusRequestEntityTooLarge, "%s", runtime.ErrSourceLimit)
	}

	languages := sandbox.Template.Language
	data["Sandbox"] = dto.ToSandbox(sandbox)
	data["Language"] = selectLanguage(languages, ctx.Query("l"))
	data["Languages"] = languages
	data["Code"] = code

	t.HTML(http.StatusOK, "sandbox")
	return nil
}

// Execute execute code in a sandbox.
// @Summary Execute code in a sandbox
// @Tags Runner
// @Description Uses the sandbox template permissions and limits. Missing or unsupported lang selects the first configured language. Unknown sandbox UIDs redirect to /.
// @Accept x-www-form-urlencoded
// @Produce json
// @Param uid path string true "UID"
// @Param lang formData string false "Programming language"
// @Param code formData string false "Source code"
// @Success 200 {object} dto.Response[ExecutionResult]
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 413 {object} dto.ErrorResponse "Request body or source exceeds configured limits"
// @Failure 429 {object} dto.ErrorResponse "Per-client task limit reached"
// @Failure 503 {object} dto.ErrorResponse "Global task slots and queue are full, or the request body memory budget is exhausted"
// @Success 302 "Sandbox not found; redirect to /"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID executeCode
// @Router /r/{uid}/execute [post]
func (h *RunnerHandler) Execute(ctx context.Context, sandbox *db.Sandbox) error {
	if err := ctx.Request().ParseForm(); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to parse form")
		return ctx.Error(http.StatusBadRequest, "Failed to parse form: %v", err)
	}

	startAt := time.Now().UnixNano()

	r, err := runtime.NewTask(ctx.Request().Context(), runtime.NewTaskOptions{
		Language: selectLanguage(sandbox.Template.Language, ctx.Request().PostForm.Get("lang")),
		Template: sandbox.Template,
		Code:     []byte(ctx.Request().PostForm.Get("code")),
		Client:   clientAddress(ctx.Request().Request),
	})
	if err != nil {
		if errors.Is(err, runtime.ErrSourceLimit) {
			return ctx.Error(http.StatusRequestEntityTooLarge, "%s", err)
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create task")
		return ctx.Error(http.StatusInternalServerError, "Failed to create task: %v", err)
	}

	output, err := r.Run(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, runtime.ErrClientTaskLimit) {
			return ctx.Error(http.StatusTooManyRequests, "%s", err)
		}
		if errors.Is(err, runtime.ErrTaskLimit) {
			return ctx.Error(http.StatusServiceUnavailable, "%s", err)
		}
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to run task")
		return ctx.Error(http.StatusInternalServerError, "Failed to run task: %v", err)
	}

	return ctx.Success(ExecutionResult{
		Result:  output,
		StartAt: startAt,
		EndAt:   time.Now().UnixNano(),
	})
}
