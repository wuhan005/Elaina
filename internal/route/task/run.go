package task

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wuhan005/gadget"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/ratelimit"
	"github.com/wuhan005/Elaina/internal/task"
)

func SandboxMiddleware(c *gin.Context) {
	uid := c.Param("uid")
	sandbox, err := db.Sandboxes.GetByUID(uid)
	if err != nil {
		c.HTML(404, "sandbox_404.tmpl", nil)
		c.Abort()
		return
	}
	c.Set("uid", uid)
	c.Set("sandbox", sandbox)
	c.Next()
}

func EditorHandler(c *gin.Context) {
	sandboxIf := c.MustGet("sandbox")
	sandbox, ok := sandboxIf.(*db.Sandbox)
	if !ok {
		c.HTML(404, "sandbox_404.tmpl", nil)
		return
	}

	var selectLang string
	var lang []string
	for _, l := range sandbox.Template.Language.Elements {
		if l.String == c.Query("l") {
			selectLang = l.String
		}
		lang = append(lang, l.String)
	}

	if len(lang) == 0 {
		c.HTML(404, "sandbox_404.tmpl", nil)
		return
	}

	// Not setting languages
	if selectLang == "" {
		selectLang = lang[0]
	}

	code := c.PostForm("c")
	if code == "" {
		code = sandbox.Placeholder
	}

	c.HTML(200, "sandbox.tmpl", gin.H{
		"Sandbox":   sandbox,
		"Language":  selectLang,
		"Languages": lang,
		"Code":      code,
	})
}

func RunTaskHandler(c *gin.Context) (int, interface{}) {
	selectLang := c.PostForm("lang")
	code := c.PostForm("code")

	sandboxIf := c.MustGet("sandbox")
	sandbox, ok := sandboxIf.(*db.Sandbox)
	if !ok {
		return gadget.MakeErrJSON(50000, "Failed to get sandbox data.")
	}

	var lang []string
	for _, l := range sandbox.Template.Language.Elements {
		lang = append(lang, l.String)
	}

	if len(lang) == 0 {
		return gadget.MakeErrJSON(50000, "Failed to get language")
	}

	// Not setting languages
	if selectLang == "" {
		selectLang = lang[0]
	}

	// Rete limit
	templateRateKey := fmt.Sprintf("tpl-%d", sandbox.TemplateID)

	// ⚠️ Using `X-Forwarded-For`, `X-Real-Ip`, `X-Appengine-Remote-Addr` as the IP address at first.
	// If you deploy Elaina without a reverse proxy server, you should pay attention to it.
	// The attackers can bypass the IP rate limit by changing the HTTP headers.
	ipRateKey := fmt.Sprintf("ip-%s", c.ClientIP())

	err := ratelimit.Add(templateRateKey, sandbox.Template.MaxContainer)
	if err != nil {
		return gadget.MakeErrJSON(40300, "rate limit: max container limit.")
	}
	defer ratelimit.Done(templateRateKey)

	err = ratelimit.Add(ipRateKey, sandbox.Template.MaxContainerPerIP)
	if err != nil {
		return gadget.MakeErrJSON(40300, "rate limit: ip limit.")
	}
	defer ratelimit.Done(ipRateKey)

	startAt := time.Now().UnixNano()

	var t task.Runner
	if os.Getenv("ELAINA_KUBERNETES_MODE") == "on" {
		t, err = task.NewKubernetesTask(c, task.NewKubernetesTaskOptions{
			Language: selectLang,
			Template: sandbox.Template,
			Code:     []byte(code),
		})
	} else {
		t, err = task.NewDockerTask(c, task.NewDockerTaskOptions{
			Language: selectLang,
			Template: sandbox.Template,
			Code:     []byte(code),
		})
	}
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to create task: %v", err)
	}

	output, err := t.Run(c)
	if err != nil {
		return gadget.MakeErrJSON(50000, "Failed to run task: %v", err)
	}

	endAt := time.Now().UnixNano()

	return gadget.MakeSuccessJSON(gin.H{
		"result":   output,
		"start_at": startAt,
		"end_at":   endAt,
	})
}
