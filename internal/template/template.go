package template

type Template struct {
	Exclude []string
}

var Templates map[string]Template

var resticTemplate Template = Template{
	Exclude: []string{"config"},
}

var macOSTemplate Template = Template{
	Exclude: []string{".DS_Store"},
}

func init() {
	Templates = map[string]Template{
		"restic": resticTemplate,
		"darwin": macOSTemplate,
	}
}
