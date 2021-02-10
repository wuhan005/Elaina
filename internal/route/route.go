package route

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

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

	run := r.Group("/r")
	{
		run.GET("/", task.RunTaskHandler)
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

	r.StaticFile("/static", "./public")
	return r
}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
