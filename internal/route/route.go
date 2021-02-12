package route

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/wuhan005/Elaina/internal/route/sandbox"
	"github.com/wuhan005/Elaina/internal/route/task"
	"github.com/wuhan005/Elaina/internal/route/template"
)

// New returns a new gin router.
func New() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{"Authorization", "Content-type", "User-Agent"},
		AllowOrigins: []string{"*"},
	}))

	r.LoadHTMLGlob("templates/*")

	run := r.Group("/r")
	run.Use(task.SandboxMiddleware)
	{
		run.GET("/:uid", task.EditorHandler)
		run.POST("/:uid", __(task.RunTaskHandler))
	}

	api := r.Group("/api")
	managerApi := api.Group("/m")
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
	// /fe will be crated by CI.
	r.Static("/m", "./fe")

	r.Static("/static", "./public")
	return r
}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
