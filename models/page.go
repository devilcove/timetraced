package models

type Page struct {
	Page               string
	Version            string
	Theme              string
	Font               string
	Tracking           bool
	Summary            map[string]string
	CurrentProject     string
	CurrentSession     string
	CurrentProjectTime string
	Today              string
}

var page Page

func init() {
	page.Version = "v0.1.0"
	page.Theme = "indigo"
	page.Font = "Roboto"
}

func GetPage() Page {
	return page
}

func SetTheme(theme string) {
	page.Theme = theme
}

func SetFont(font string) {
	page.Font = font
}
