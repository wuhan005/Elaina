// Copyright 2024 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	gocontext "context"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/flamego/flamego"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/dto"
	"github.com/wuhan005/Elaina/internal/languages"
	"github.com/wuhan005/Elaina/internal/runtime"
	"github.com/wuhan005/Elaina/public"
	"github.com/wuhan005/Elaina/templates"
)

const sandboxScriptPath = "js/sandbox.js"

type RunnerHandler struct {
	page       *template.Template
	script     []byte
	codeMirror codeMirrorAssets
}

type ExecutionResult struct {
	Result  *runtime.ExecOutput `json:"result"`
	StartAt int64               `json:"start_at"`
	EndAt   int64               `json:"end_at"`
} // @name ExecutionResult

func NewRunnerHandler() (*RunnerHandler, error) {
	m := minify.New()
	m.AddFunc("application/javascript", js.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
		KeepQuotes:       true,
		TemplateDelims:   html.GoTemplateDelims,
	})

	// Minify the template before rendering so source code entered by users is never minified.
	const pageName = "sandbox.tmpl"
	pageSource, err := templates.FS.ReadFile(pageName)
	if err != nil {
		return nil, errors.Wrap(err, "read sandbox template")
	}

	pageSource, err = m.Bytes("text/html", pageSource)
	if err != nil {
		return nil, errors.Wrap(err, "minify sandbox template")
	}

	page, err := template.New(pageName).Parse(string(pageSource))
	if err != nil {
		return nil, errors.Wrap(err, "parse sandbox template")
	}

	script, err := public.FS.ReadFile(sandboxScriptPath)
	if err != nil {
		return nil, errors.Wrap(err, "read sandbox script")
	}

	script, err = m.Bytes("application/javascript", script)
	if err != nil {
		return nil, errors.Wrap(err, "minify sandbox script")
	}

	return &RunnerHandler{
		page:       page,
		script:     script,
		codeMirror: newCodeMirrorAssets(config.App.CodeMirror),
	}, nil
}

type codeMirrorAssets struct {
	CSS     string
	Scripts []string
}

func newCodeMirrorAssets(source config.CodeMirrorOptions) codeMirrorAssets {
	core, suffix := source.BaseURL, ".min"
	if source.Layout == config.CodeMirrorLayoutNPM {
		core += "lib/"
		suffix = ""
	}

	assets := codeMirrorAssets{
		CSS:     core + "codemirror" + suffix + ".css",
		Scripts: []string{core + "codemirror" + suffix + ".js"},
	}
	for _, mode := range []string{"xml", "htmlmixed", "clike", "php", "python", "go", "javascript"} {
		assets.Scripts = append(assets.Scripts, source.BaseURL+"mode/"+mode+"/"+mode+suffix+".js")
	}

	return assets
}

func (h *RunnerHandler) serveScript(ctx flamego.Context) {
	r := ctx.Request().Request
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return
	}
	if strings.TrimRight(r.URL.Path, "/") != "/static/"+sandboxScriptPath {
		return
	}

	http.ServeContent(ctx.ResponseWriter(), r, sandboxScriptPath, time.Time{}, bytes.NewReader(h.script))
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

type executionRecord struct {
	request  *http.Request
	sandbox  *db.Sandbox
	language string
	code     string
}

// save stores the execution; rejected requests keep only the code size so floods cannot fill the database.
func (e executionRecord) save(status db.ExecutionStatus, output *runtime.ExecOutput, runErr error) {
	options := db.CreateExecutionOptions{
		SandboxID:     e.sandbox.ID,
		Language:      e.language,
		Status:        status,
		CodeBytes:     len(e.code),
		ClientAddress: "unknown",
		UserAgent:     e.request.UserAgent(),
	}
	if addr, ok := clientIP(e.request); ok {
		options.ClientAddress = addr.WithZone("").String()
	}
	if status != db.ExecutionStatusRejected {
		options.Code = e.code
	}
	if output != nil {
		options.Stdout, options.Stderr, options.Error, options.Duration = output.Stdout, output.Stderr, output.Error, output.Duration
	}
	if runErr != nil {
		options.Error = runErr.Error()
	}

	// Still save when the client has already disconnected.
	ctx, cancel := gocontext.WithTimeout(gocontext.WithoutCancel(e.request.Context()), 5*time.Second)
	defer cancel()
	if _, err := db.Executions.Create(ctx, options); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to save execution")
	}
}

// selectLanguage returns language if the template supports it, otherwise the first supported language.
func selectLanguage(languages []string, language string) string {
	if lo.Contains(languages, language) {
		return language
	}
	return languages[0]
}

func (h *RunnerHandler) View(ctx context.Context, sandbox *db.Sandbox) error {
	if err := ctx.Request().ParseForm(); err != nil {
		return ctx.Error(http.StatusBadRequest, "runner::parse_form")
	}

	code := ctx.Request().PostForm.Get("c")
	if code == "" {
		code = sandbox.Placeholder
	}
	if int64(len(code)) > config.App.WASMLimits.MaxSourceBytes {
		return ctx.Error(http.StatusRequestEntityTooLarge, "runner::source_too_large")
	}

	names := sandbox.Template.Language
	language := selectLanguage(names, ctx.Query("l"))
	options := make([]languageOption, 0, len(names))
	for _, name := range names {
		label := name
		if runner, ok := languages.Get(name); ok {
			label = runner.Label
		}
		options = append(options, languageOption{Value: name, Label: label})
	}

	text := runnerPageText(ctx)
	data := map[string]any{
		"Lang":            ctx.Tr("runner_page::html_lang"),
		"Sandbox":         dto.ToSandbox(sandbox),
		"Language":        language,
		"LanguageOptions": options,
		"Editable":        sandbox.Editable,
		"Code":            code,
		"Text":            text,
		"CodeMirror":      h.codeMirror,
		"Config": map[string]any{
			"language": language,
			"editable": sandbox.Editable,
			"text":     text,
		},
	}

	var page bytes.Buffer
	w := ctx.ResponseWriter()
	if err := h.page.Execute(&page, data); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to render sandbox page")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := page.WriteTo(w); err != nil {
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to write sandbox page")
	}
	return nil
}

type languageOption struct {
	Value string
	Label string
}

// runnerPageText is the copy of the runner page in the request locale; the page script uses it too.
func runnerPageText(ctx context.Context) map[string]string {
	return map[string]string{
		"run":                ctx.Tr("runner_page::run"),
		"running":            ctx.Tr("runner_page::running"),
		"open_in_new_window": ctx.Tr("runner_page::open_in_new_window"),
		"output":             ctx.Tr("runner_page::output"),
		"clear":              ctx.Tr("runner_page::clear"),
		"idle":               ctx.Tr("runner_page::idle"),
		"read_only":          ctx.Tr("runner_page::read_only"),
		"success":            ctx.Tr("runner_page::success"),
		"failed":             ctx.Tr("runner_page::failed"),
		"timed_out":          ctx.Tr("runner_page::timed_out"),
		"request_failed":     ctx.Tr("runner_page::request_failed"),
		"no_output":          ctx.Tr("runner_page::no_output"),
		"stderr":             ctx.Tr("runner_page::stderr"),
		"error":              ctx.Tr("runner_page::error"),
		"language":           ctx.Tr("runner_page::language"),
		"resize":             ctx.Tr("runner_page::resize"),
		"powered_by":         ctx.Tr("runner_page::powered_by"),
	}
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
		return ctx.Error(http.StatusBadRequest, "runner::parse_form_detail", err)
	}

	startAt := time.Now().UnixNano()
	execution := executionRecord{
		request:  ctx.Request().Request,
		sandbox:  sandbox,
		language: selectLanguage(sandbox.Template.Language, ctx.Request().PostForm.Get("lang")),
		code:     ctx.Request().PostForm.Get("code"),
	}

	r, err := runtime.NewTask(ctx.Request().Context(), runtime.NewTaskOptions{
		Language: execution.language,
		Template: sandbox.Template,
		Code:     []byte(execution.code),
		Client:   clientAddress(ctx.Request().Request),
	})
	if err != nil {
		if errors.Is(err, runtime.ErrSourceLimit) {
			execution.save(db.ExecutionStatusRejected, nil, err)
			return ctx.Error(http.StatusRequestEntityTooLarge, "runner::source_too_large")
		}
		execution.save(db.ExecutionStatusError, nil, err)
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to create task")
		return ctx.Error(http.StatusInternalServerError, "runner::create_task_failed", err)
	}

	output, err := r.Run(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, runtime.ErrClientTaskLimit) {
			execution.save(db.ExecutionStatusRejected, nil, err)
			return ctx.Error(http.StatusTooManyRequests, "runner::client_task_limit")
		}
		if errors.Is(err, runtime.ErrTaskLimit) {
			execution.save(db.ExecutionStatusRejected, nil, err)
			return ctx.Error(http.StatusServiceUnavailable, "runner::task_limit")
		}
		execution.save(db.ExecutionStatusError, nil, err)
		logrus.WithContext(ctx.Request().Context()).WithError(err).Error("Failed to run task")
		return ctx.Error(http.StatusInternalServerError, "runner::run_task_failed", err)
	}

	switch output.Error {
	case "":
		execution.save(db.ExecutionStatusSuccess, output, nil)
	case runtime.ExecErrorTimeout:
		execution.save(db.ExecutionStatusTimeout, output, nil)
	default:
		execution.save(db.ExecutionStatusFailed, output, nil)
	}

	return ctx.Success(ExecutionResult{
		Result:  output,
		StartAt: startAt,
		EndAt:   time.Now().UnixNano(),
	})
}
