package views

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

var (
	// layoutDir
	layoutDir = "views/layouts/"
	// templateDir is used to prepend a path to view templates
	templateDir = "views/"

	// templateExt is template files extention
	templateExt = ".gohtml"
)

// NewView creates a template for the new view and parses necessary files
func NewView(layout string, files ...string) *View {
	files = append(files, func() []string {
		f, err := filepath.Glob(layoutDir + "*" + templateExt)
		if err != nil {
			log.Fatal(err)
		}
		// for Windows build
		for i := range f {
			f[i] = filepath.ToSlash(f[i])
		}
		return f
	}()...)
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}
	return &View{
		Template: tpl,
		Layout:   layout,
	}
}

// View represents a web page view
type View struct {
	Layout   string
	Template *template.Template
}

type ViewData struct {
	MeetingActive bool
	GroupName     string
	MeetingTitle  string
	Data          interface{}
}

// Render creates an image of a web page from templates and layouts and writes it to the response
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	var vd ViewData
	switch d := data.(type) {
	case ViewData:
		vd = d
	default:
		vd = ViewData{
			Data: data,
		}
	}
	w.Header().Set("Content-type", "text/html")
	var buf bytes.Buffer
	if err := v.Template.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		log.Printf("views:render: %v", err)
		http.Error(w, "Sorry, something went wrong!", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("views:render: %v", err)
		http.Error(w, "Sorry, something went wrong!", http.StatusInternalServerError)
	}
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}
