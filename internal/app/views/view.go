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
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)

	t, err := template.New("").ParseFiles(files...)
	if err != nil {
		log.Fatal(err)
	}

	return &View{
		t,
		layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var vd Data

	switch d := data.(type) {
	case Data:
		vd = d
		// do nothing
	default:
		vd = Data{
			Yield: data,
		}
	}
	if alert := getAlert(r); alert != nil {
		vd.Alert = alert
		clearAlert(w)
	}

	var buf bytes.Buffer
	var err error

	err = v.Template.ExecuteTemplate(&buf, v.Layout, vd)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, &buf)
	if err != nil {
		log.Println(err)
	}
}

// layoutFiles return slice of string represents list of files used in app.
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return files
}

// addTemplatePath : "index" -> "/views/index"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt : "index" -> "index.gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}
