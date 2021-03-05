package route

import (
	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {
	c.HTML(200, "index.tmpl", gin.H{})
}
