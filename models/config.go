package models

// Config represents the user selectable UI options.
type Config struct {
	Theme   string `form:"theme"   json:"theme"`
	Font    string `form:"font"    json:"font"`
	Refresh int    `form:"refresh" json:"refresh"`
}
