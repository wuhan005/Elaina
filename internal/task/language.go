package task

type runner struct {
	Name  string
	Ext   string
	Image string
	Cmd   []string
}

var langRunners = []runner{
	{
		Name:  "PHP",
		Ext:   ".php",
		Image: "elaina-php:latest",
		Cmd:   []string{"sh", "-c", "php /runner/code.php"},
	},
}
