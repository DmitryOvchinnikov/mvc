package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dmitryovchinnikov/mvc/internal/app/models"
	"github.com/dmitryovchinnikov/mvc/internal/app/views"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	ShowTable = "show_table"
	EditTable = "edit_table"

	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewTables(s models.TableService, r *mux.Router) *Tables {
	return &Tables{
		New:       views.NewView("bootstrap", "new"),
		IndexView: views.NewView("bootstrap", "index"),
		EditView:  views.NewView("bootstrap", "edit"),
		service:   s,
		router:    r,
	}
}

type Tables struct {
	New       *views.View
	ShowView  *views.View
	IndexView *views.View
	EditView  *views.View
	service   models.TableService
	router    *mux.Router
}

type TableForm struct {
	Text   string    `schema:"text"`
	CodeID uuid.UUID `schema:"code_id"`
}

// Index
// GET /index
func (t *Tables) Index(w http.ResponseWriter, r *http.Request) {
	tables, err := t.service.SelectAll()
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	var vd views.Data
	vd.Yield = tables
	t.IndexView.Render(w, r, vd)
}

// Show
// GET /index/:id
func (t *Tables) Show(w http.ResponseWriter, r *http.Request) {
	tbl, err := t.tablesByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = tbl
	t.ShowView.Render(w, r, vd)
}

// Edit
// GET /index/:id/edit
func (t *Tables) Edit(w http.ResponseWriter, r *http.Request) {
	tbl, err := t.tablesByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = tbl
	t.EditView.Render(w, r, vd)
}

// Update
// POST /index/:id/update
func (t *Tables) Update(w http.ResponseWriter, r *http.Request) {
	var err error

	tbl, err := t.tablesByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = tbl

	var form TableForm
	err = parseForm(r, &form)
	if err != nil {
		vd.SetAlert(err)
		t.EditView.Render(w, r, vd)
		return
	}

	tbl.Text = form.Text
	tbl.CodeID = form.CodeID
	err = t.service.Update(tbl)
	if err != nil {
		vd.SetAlert(err)
		t.EditView.Render(w, r, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLevelSuccess,
		Message: "Table successfully updated!",
	}
	t.EditView.Render(w, r, vd)
}

// Create
// POST /tables
func (t *Tables) Create(w http.ResponseWriter, r *http.Request) {
	var err error

	var vd views.Data
	var form TableForm
	err = parseForm(r, &form)
	if err != nil {
		vd.SetAlert(err)
		t.New.Render(w, r, vd)
		return
	}

	tbl := models.Table{
		Text:   form.Text,
		CodeID: form.CodeID,
	}
	err = t.service.Create(&tbl)
	if err != nil {
		vd.SetAlert(err)
		t.New.Render(w, r, vd)
		return
	}

	url, err := t.router.Get(EditTable).URL("id", fmt.Sprintf("%v", tbl.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/tables", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// TextUpload
// POST /tables/:id/text
func (t *Tables) TextUpload(w http.ResponseWriter, r *http.Request) {
	var err error

	tbl, err := t.tablesByID(w, r)
	if err != nil {
		return
	}

	var vd views.Data
	vd.Yield = tbl
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		t.EditView.Render(w, r, vd)
		return
	}

	url, err := t.router.Get(EditTable).URL("id", fmt.Sprintf("%v", tbl.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/tables", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (t *Tables) tablesByID(w http.ResponseWriter, r *http.Request) (*models.Table, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid table ID...", http.StatusNotFound)
		return nil, err
	}

	tbl, err := t.service.ByID(id)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Table not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Whoops! Something went wrong...", http.StatusInternalServerError)
		}
		return nil, err
	}

	return tbl, nil
}
