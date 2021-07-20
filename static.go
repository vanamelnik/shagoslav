package shagoslav

import "shagoslav/views"

func NewStatic() *Static {
	return &Static{
		Home: views.NewView("simple", "views/static/index.gohtml"),
	}
}

// Static represents static pages controller
type Static struct {
	Home *views.View
}
