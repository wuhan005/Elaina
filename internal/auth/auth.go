package auth

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/wuhan005/gadget"
)

const loginFlag = "i love you"

func LoginMiddleware(c *gin.Context) {
	if c.Request.RequestURI == "/api/m/login" || c.Request.RequestURI == "/api/m/logout" {
		c.Next()
		return
	}
	session := sessions.Default(c)
	sess, ok := session.Get("elaina").(string)
	if !ok || sess != loginFlag {
		c.JSON(400, gin.H{
			"error": 40000,
			"msg":   "Auth required",
		})
		c.Abort()
		return
	}
	c.Next()
}

func LoginHandler(c *gin.Context) (int, interface{}) {
	var form struct {
		Password string `json:"password"`
	}
	err := c.BindJSON(&form)
	if err != nil {
		return gadget.MakeErrJSON(40000, "Failed to bind JSON.")
	}
	if form.Password != os.Getenv("APP_PASSWORD") {
		return gadget.MakeErrJSON(40100, "Wrong password.")
	}

	session := sessions.Default(c)
	session.Set("elaina", loginFlag)
	err = session.Save()
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to save session data.")
	}

	return gadget.MakeSuccessJSON("Login succeed.")
}

func LogoutHandler(c *gin.Context) (int, interface{}) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	return gadget.MakeSuccessJSON("Logout succeed.")
}

func CheckStatusHandlers(c *gin.Context) (int, interface{}) {
	return gadget.MakeSuccessJSON("")
}
