package main

type Template struct {
	Exclude []string
}

var templates map[string]Template

var resticTemplate Template = Template{
	Exclude: []string{"config"},
}

var macOSTemplate Template = Template{
	Exclude: []string{".DS_Store"},
}

func init() {
	templates = map[string]Template{
		"restic": resticTemplate,
		"darwin": macOSTemplate,
	}
}
