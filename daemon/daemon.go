package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/wuhan005/gadget"
	log "unknwon.dev/clog/v2"
)

func main() {
	_ = log.NewConsole()
	defer log.Stop()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		hostName, _ := os.Hostname()
		c.JSON(gadget.MakeSuccessJSON(hostName))
	})

	// Create file.
	r.POST("/create", func(c *gin.Context) {
		var input struct {
			FileName string `json:"file_name"`
			Data     string `json:"data"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			c.JSON(gadget.MakeErrJSON(40000, "Daemon: unexpected input"))
			return
		}

		err = ioutil.WriteFile(input.FileName, []byte(input.Data), 0644)
		if err != nil {
			c.JSON(gadget.MakeErrJSON(400, err.Error()))
			return
		}
		c.JSON(gadget.MakeSuccessJSON("success"))
	})

	// Execute command.
	r.POST("/exec", func(c *gin.Context) {
		var rawCommand string
		err := c.BindJSON(&rawCommand)
		if err != nil {
			c.JSON(gadget.MakeErrJSON(40000, "Daemon: unexpected input"))
			return
		}

		stdOut := bytes.NewBuffer(nil)
		stdErr := bytes.NewBuffer(nil)

		cmd := exec.Command("/bin/sh", "-c", rawCommand)
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		err = cmd.Run()

		c.JSON(gadget.MakeSuccessJSON(gin.H{
			"stdout": stdOut.String(),
			"stderr": stdErr.String(),
			"error":  err != nil,
		}))
	})

	err := r.RunUnix("./elaina-daemon.sock")
	if err != nil {
		log.Fatal("Failed to start daemon: %v", err)
	}
}
