package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

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
		c.JSON(http.StatusOK, gin.H{"host_name": hostName})
	})

	// Create file.
	r.POST("/create", func(c *gin.Context) {
		var input struct {
			FileName string `json:"file_name"`
			Data     string `json:"data"`
		}
		err := c.BindJSON(&input)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error":  true,
				"stderr": "Daemon: unexpected input",
			})
			return
		}

		err = ioutil.WriteFile(path.Join("/runtime/runner", input.FileName), []byte(input.Data), 0644)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error":  true,
				"stderr": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"error": false})
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
		cmd.Dir = "/runtime/runner"
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		err = cmd.Run()

		c.JSON(http.StatusOK, gin.H{
			"error":  err != nil,
			"stdout": stdOut.String(),
			"stderr": stdErr.String(),
		})
	})

	err := r.RunUnix("/runtime/elaina-daemon.sock")
	if err != nil {
		log.Fatal("Failed to start daemon: %v", err)
	}
}
