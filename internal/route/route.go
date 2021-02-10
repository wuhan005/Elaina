package route

import (
	"github.com/gin-gonic/gin"

	"github.com/wuhan005/Elaina/internal/route/task"
	"github.com/wuhan005/Elaina/internal/route/template"
)

// New returns a new gin router.
func New() *gin.Engine {
	r := gin.Default()

	run := r.Group("/r")
	{
		run.GET("/", task.RunTaskHandler)
	}

	manager := r.Group("/m")
	{
		manager.GET("/templates", __(template.ListTemplatesHandler))
		manager.GET("/template", __(template.GetTemplateHandler))
		manager.POST("/template", __(template.CreateTemplateHandler))
		manager.PUT("/template", __(template.UpdateTemplateHandler))
		manager.DELETE("/template", __(template.DeleteTemplateHandler))
	}

	r.StaticFile("/static", "./public")
	return r
}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
