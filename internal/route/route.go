package route

import (
	"github.com/gin-gonic/gin"

	"github.com/wuhan005/Elaina/internal/route/task"
)

// New returns a new gin router.
func New() *gin.Engine {
	r := gin.Default()

	r.GET("")

	run := r.Group("/run")
	{
		run.GET("/", task.RunTaskHandler)
	}

	return r
}
