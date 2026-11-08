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

	data["Sandbox"] = dto.ToSandbox(sandbox)
	data["Language"] = selectedLanguage
	data["Languages"] = languages
	data["Code"] = code

	t.HTML(http.StatusOK, "sandbox")
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
// @Success 302 "Sandbox not found; redirect to /"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @ID executeCode
// @Router /r/{uid}/execute [post]
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

	r, err := runtime.NewTask(ctx.Request().Context(), runtime.NewTaskOptions{
		Language: selectedLanguage,
		Template: sandbox.Template,
		Code:     []byte(code),
	})
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

	return ctx.Success(ExecutionResult{
		Result:  output,
		StartAt: startAt,
		EndAt:   endAt,
	})
}
