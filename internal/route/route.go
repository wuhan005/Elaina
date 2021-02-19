package route

import (
	tmpl "html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/frontend"
	"github.com/wuhan005/Elaina/internal/auth"
	"github.com/wuhan005/Elaina/internal/route/sandbox"
	"github.com/wuhan005/Elaina/internal/route/task"
	"github.com/wuhan005/Elaina/internal/route/template"
	"github.com/wuhan005/Elaina/public"
	"github.com/wuhan005/Elaina/templates"
)

// New returns a new gin router.
func New() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Content-type", "User-Agent"},
		AllowCredentials: true,
		AllowOrigins:     []string{os.Getenv("APP_URL")},
	}))

	// Session
	store := cookie.NewStore([]byte(randstr.String(50)))
	r.Use(sessions.Sessions("elaina", store))

	// Templates
	tpl := tmpl.Must(tmpl.New("").ParseFS(templates.FS, "*"))
	r.SetHTMLTemplate(tpl)

	run := r.Group("/r")
	run.Use(task.SandboxMiddleware)
	{
		run.GET("/:uid", task.EditorHandler)
		run.POST("/:uid", __(task.RunTaskHandler))
	}

	api := r.Group("/api")
	managerApi := api.Group("/m")
	managerApi.Use(auth.LoginMiddleware)
	{
		managerApi.POST("/login", __(auth.LoginHandler))
		managerApi.POST("/logout", __(auth.LogoutHandler))
		managerApi.GET("/status", __(auth.CheckStatusHandlers))
	}
	{
		managerApi.GET("/templates", __(template.ListTemplatesHandler))
		managerApi.GET("/template", __(template.GetTemplateHandler))
		managerApi.POST("/template", __(template.CreateTemplateHandler))
		managerApi.PUT("/template", __(template.UpdateTemplateHandler))
		managerApi.DELETE("/template", __(template.DeleteTemplateHandler))
	}
	{
		managerApi.GET("/sandboxes", __(sandbox.ListSandboxesHandler))
		managerApi.GET("/sandbox", __(sandbox.GetSandboxHandler))
		managerApi.POST("/sandbox", __(sandbox.CreateSandboxHandler))
		managerApi.PUT("/sandbox", __(sandbox.UpdateSandboxHandler))
		managerApi.DELETE("/sandbox", __(sandbox.DeleteSandboxHandler))
	}
	// /m will be created by CI.
	fe, err := fs.Sub(frontend.FS, "dist")
	if err != nil {
		log.Fatal("Failed to sub path `dist`: %v", err)
	}
	r.StaticFS("/m", http.FS(fe))
	r.StaticFS("/static", http.FS(public.FS))
	return r
}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
