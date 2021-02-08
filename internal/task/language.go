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
}
