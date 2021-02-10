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
	manager := api.Group("/m")
	{
		manager.GET("/templates", __(template.ListTemplatesHandler))
		manager.GET("/template", __(template.GetTemplateHandler))
		manager.POST("/template", __(template.CreateTemplateHandler))
		manager.PUT("/template", __(template.UpdateTemplateHandler))
		manager.DELETE("/template", __(template.DeleteTemplateHandler))
	}
	{
		manager.GET("/sandboxes", __(sandbox.ListSandboxesHandler))
		manager.GET("/sandbox", __(sandbox.GetSandboxHandler))
		manager.POST("/sandbox", __(sandbox.CreateSandboxHandler))
		manager.PUT("/sandbox", __(sandbox.UpdateTemplateHandler))
		manager.DELETE("/sandbox", __(sandbox.DeleteTemplateHandler))
	}

	r.Static("/static", "./public")
	return r
}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
