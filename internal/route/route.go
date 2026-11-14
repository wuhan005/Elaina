package route

import (
	"io/fs"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/template"
	"github.com/pkg/errors"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/form"
	"github.com/wuhan005/Elaina/internal/i18n"
	"github.com/wuhan005/Elaina/public"
	"github.com/wuhan005/Elaina/templates"
	"github.com/wuhan005/Elaina/web"
)

// New returns a new Flamego router.
// @Title Elaina API
// @Version 1.0
// @Description Management and WASM code execution API. Successful JSON responses use a data envelope; errors contain msg. All HTTP bodies are capped by app.wasm_limits.max_request_bytes; oversized bodies return HTTP 413. Buffered bodies share app.wasm_limits.max_buffered_request_bytes; requests that exceed it return HTTP 503.
// @BasePath /
// @SecurityDefinitions.apikey SessionAuth
// @In header
// @Name Authorization
// @Description Use "Token <session token>" with the token returned by signIn.
func New() (*flamego.Flame, error) {
	f := flamego.Classic()

	frontendFS, err := fs.Sub(web.FS, "dist")
	if err != nil {
		return nil, errors.Wrap(err, "fs sub")
	}

	templatesFS, err := template.EmbedFS(templates.FS, ".", []string{".tmpl"})
	if err != nil {
		return nil, errors.Wrap(err, "embed templates fs")
	}

	f.Use(
		limitRequestBody,
		session.Sessioner(session.Options{
			// TODO: support postgresql
			Initer:      session.MemoryIniter(),
			ReadIDFunc:  context.ReadIDFunc,
			WriteIDFunc: context.WriteIDFunc,
		}),

		// Public static files.
		flamego.Static(flamego.StaticOptions{
			FileSystem: http.FS(public.FS),
			Prefix:     "static",
		}),
		// Frontend static files.
		flamego.Static(flamego.StaticOptions{
			FileSystem: http.FS(frontendFS),
		}),
		template.Templater(template.Options{
			FileSystem: templatesFS,
		}),
		i18n.I18n(),
		context.Contexter(),
	)

	runnerHandler := NewRunnerHandler()
	f.Group("/r/{uid}", func() {
		f.Combo("").Get(runnerHandler.View).Post(runnerHandler.View)
		f.Post("/execute", runnerHandler.Execute)
	}, runnerHandler.Runner)

	f.Group("/api", func() {
		authHandler := NewAuthHandler()
		f.Post("/auth/sign-in", form.Bind(form.SignIn{}), authHandler.SignIn)
		f.Post("/auth/sign-in/2fa", form.Bind(form.SignInTwoFactor{}), authHandler.SignInTwoFactor)

		f.Group("", func() {
			f.Group("/auth", func() {
				f.Post("/sign-out", authHandler.SignOut)
				f.Get("/profile", authHandler.Profile)
			})

			accountHandler := NewAccountHandler()
			f.Group("/account", func() {
				f.Put("/password", form.Bind(form.ChangePassword{}), accountHandler.ChangePassword)
				f.Post("/2fa/setup", accountHandler.SetupTwoFactor)
				f.Post("/2fa/enable", form.Bind(form.EnableTwoFactor{}), accountHandler.EnableTwoFactor)
				f.Post("/2fa/disable", form.Bind(form.ConfirmPassword{}), accountHandler.DisableTwoFactor)
				f.Post("/2fa/recovery-codes", form.Bind(form.ConfirmPassword{}), accountHandler.RegenerateRecoveryCodes)
			})

			executionHandler := NewExecutionHandler()
			f.Get("/executions", executionHandler.List)
			f.Get("/executions/stats", executionHandler.Stats)
			f.Get("/execution/{id}", executionHandler.Executioner, executionHandler.Get)

			userHandler := NewUserHandler()
			f.Combo("/users").
				Get(userHandler.List).
				Post(form.Bind(form.CreateUser{}), userHandler.Create)
			f.Group("/user/{id}", func() {
				f.Delete("", userHandler.Delete)
				f.Put("/password", form.Bind(form.ResetPassword{}), userHandler.ResetPassword)
				f.Delete("/2fa", userHandler.ResetTwoFactor)
			}, userHandler.Userer)

			templateHandler := NewTemplateHandler()
			f.Combo("/templates").
				Get(templateHandler.List).
				Post(form.Bind(form.CreateTemplate{}), templateHandler.Create)
			f.Get("/templates/all", templateHandler.All)
			f.Group("/template/{id}", func() {
				f.Combo("").
					Get(templateHandler.Get).
					Put(form.Bind(form.UpdateTemplate{}), templateHandler.Update).
					Delete(templateHandler.Delete)
			}, templateHandler.Templater)

			sandboxHandler := NewSandboxHandler()
			f.Combo("/sandboxes").
				Get(sandboxHandler.List).
				Post(form.Bind(form.CreateSandbox{}), sandboxHandler.Create)
			f.Group("/sandbox/{id}", func() {
				f.Combo("").
					Get(sandboxHandler.Get).
					Put(form.Bind(form.UpdateSandbox{}), sandboxHandler.Update).
					Delete(sandboxHandler.Delete)
			}, sandboxHandler.Sandboxer)
		}, authHandler.Authenticator)
	})

	f.NotFound(Frontend)

	f.Get("/healthz")

	return f, nil
}
