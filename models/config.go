package models

type Config struct {
	Theme string `json:"theme" form:"theme"`
	Font  string `json:"font" form:"font"`
}
