package task

type runner struct {
	Name     string
	Ext      string
	Image    string
	BuildCmd string
	RunCmd   string
}

var langRunners = []runner{
	{
		Name:     "php",
		Ext:      ".php",
		Image:    "elainaruntime/php:latest",
		BuildCmd: "",
		RunCmd:   "php code.php",
	},
	{
		Name:     "python",
		Ext:      ".py",
		Image:    "elainaruntime/python:latest",
		BuildCmd: "",
		RunCmd:   "python3 code.py",
	},
	{
		Name:     "go",
		Ext:      ".go",
		Image:    "elainaruntime/golang:latest",
		BuildCmd: "go mod init elaina-runner && go build -v .",
		RunCmd:   "./elaina-runner",
	},
	{
		Name:     "javascript",
		Ext:      ".js",
		Image:    "elainaruntime/javascript:latest",
		BuildCmd: "",
		RunCmd:   "node code.js",
	},
}
