package route

import (
	"io/fs"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/template"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/form"
	"github.com/wuhan005/Elaina/public"
	"github.com/wuhan005/Elaina/templates"
	"github.com/wuhan005/Elaina/web"
)

// New returns a new Flamego router.
func New(db *gorm.DB) (*flamego.Flame, error) {
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
		context.Contexter(db),
	)

	runnerHandler := NewRunnerHandler()
	f.Group("/r/{uid}", func() {
		f.Combo("").Get(runnerHandler.View).Post(runnerHandler.View)
		f.Post("/execute", runnerHandler.Execute)
	}, runnerHandler.Runner)

	f.Group("/api", func() {
		authHandler := NewAuthHandler()
		f.Post("/auth/sign-in", form.Bind(form.SignIn{}), authHandler.SignIn)

		f.Group("", func() {
			f.Group("/auth", func() {
				f.Post("/sign-out", authHandler.SignOut)
				f.Get("/profile", authHandler.Profile)
			})

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
