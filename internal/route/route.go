package route

import (
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"gorm.io/gorm"

	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/form"
)

// New returns a new Flamego router.
func New(db *gorm.DB) *flamego.Flame {
	f := flamego.Classic()

	f.Use(
		session.Sessioner(session.Options{
			// TODO: support postgresql
			Initer:      session.MemoryIniter(),
			ReadIDFunc:  context.ReadIDFunc,
			WriteIDFunc: context.WriteIDFunc,
		}),

		context.Contexter(db),
	)

	// TODO Code runner templates

	baseHandler := NewBaseHandler()
	f.Get("/", baseHandler.Index)

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
			f.Combo("/sandbox/{id}", func() {
				f.Combo("").
					Get(sandboxHandler.Get).
					Put(form.Bind(form.UpdateSandbox{}), sandboxHandler.Update).
					Delete(sandboxHandler.Delete)
			}, sandboxHandler.Sandboxer)
		}, authHandler.Authenticator)
	})

	f.Get("/healthz")

	return f
}

//
//// New returns a new gin router.
//func NewA() *gin.Engine {
//	r := gin.Default()
//
//	// Session
//	store := cookie.NewStore([]byte(randstr.String(50)))
//	r.Use(sessions.Sessions("elaina", store))
//
//	r.GET("/", IndexHandler)
//
//	run := r.Group("/r")
//	run.Use(task.SandboxMiddleware)
//	{
//		run.GET("/:uid", task.EditorHandler)
//		run.POST("/:uid", task.EditorHandler)
//		run.POST("/:uid/execute", __(task.RunTaskHandler))
//	}
//
//	api := r.Group("/api")
//	managerApi := api.Group("/m")
//	managerApi.Use(auth.LoginMiddleware)
//	{
//		managerApi.POST("/login", __(auth.LoginHandler))
//		managerApi.POST("/logout", __(auth.LogoutHandler))
//		managerApi.GET("/status", __(auth.CheckStatusHandlers))
//	}
//	{
//		managerApi.GET("/templates", __(template.ListTemplatesHandler))
//		managerApi.GET("/template", __(template.GetTemplateHandler))
//		managerApi.POST("/template", __(template.CreateTemplateHandler))
//		managerApi.PUT("/template", __(template.UpdateTemplateHandler))
//		managerApi.DELETE("/template", __(template.DeleteTemplateHandler))
//	}
//	{
//		managerApi.GET("/sandboxes", __(sandbox.ListSandboxesHandler))
//		managerApi.GET("/sandbox", __(sandbox.GetSandboxHandler))
//		managerApi.POST("/sandbox", __(sandbox.CreateSandboxHandler))
//		managerApi.PUT("/sandbox", __(sandbox.UpdateSandboxHandler))
//		managerApi.DELETE("/sandbox", __(sandbox.DeleteSandboxHandler))
//	}
//
//	// /m will be created by CI.
//	fe, err := fs.Sub(frontend.FS, "dist")
//	if err != nil {
//		log.Fatal("Failed to sub path `dist`: %v", err)
//	}
//	r.StaticFS("/m", http.FS(fe))
//	r.StaticFS("/static", http.FS(public.FS))
//	r.NoRoute(func(c *gin.Context) {
//		c.Redirect(http.StatusTemporaryRedirect, "/")
//	})
//	return r
//}
//
//func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
//	return func(c *gin.Context) {
//		c.JSON(handler(c))
//	}
//}
