package task

type runner struct {
	Name  string
	Ext   string
	Image string
	Cmd   []string
}

var langRunners = []runner{
	{
		Name:  "php",
		Ext:   ".php",
		Image: "elaina-php:latest",
		Cmd:   []string{"sh", "-c", "php /runner/code.php"},
	},
	{
		Name:  "python",
		Ext:   ".py",
		Image: "elaina-python:latest",
		Cmd:   []string{"sh", "-c", "python3 /runner/code.py"},
	},
	{
		Name:  "go",
		Ext:   ".go",
		Image: "elaina-go:latest",
		Cmd:   []string{"sh", "-c", "go run /runner/code.go"},
	},
	{
		Name:  "javascript",
		Ext:   ".js",
		Image: "elaina-javascript:latest",
		Cmd:   []string{"sh", "-c", "node /runner/code.js"},
	},
}
