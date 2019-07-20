package tmplts

import (
	"fmt"
	"github.com/randallmlough/tmplts/funcmaps"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type templateNotFound interface {
	TemplateNotFound() bool
}

type errTemplateNotFound struct {
	template string
}

func (e errTemplateNotFound) Error() string {
	return fmt.Sprintf("Template '%v' not found", e.template)
}

func (e errTemplateNotFound) TemplateNotFound() bool {
	return true
}

type Templates struct {
	Templates  map[string]*template.Template
	Extensions map[string]bool

	dir         string
	stripPrefix string
	templates   []keyValue
	partials    []keyValue
	funcs       template.FuncMap
	delimsLeft  string
	delimsRight string

	// pool stores the bytes.Buffer's used when using the Render* methods
	pool *templatePool
}

type keyValue struct {
	key   string
	value string
}

func New() *Templates {
	return &Templates{
		Templates: map[string]*template.Template{},

		funcs: template.FuncMap{},

		pool: newPool(),
	}
}

func (t *Templates) Delims(left, right string) {
	t.delimsLeft = left
	t.delimsRight = right
}

func (t *Templates) ParseDir(dir string, stripPrefix string) (*Templates, error) {
	t.dir = dir
	t.stripPrefix = stripPrefix
	if err := filepath.Walk(dir, t.parseFile); err != nil {
		return nil, fmt.Errorf("templates: filepath.Walk error")
	}

	return t, nil
}

func (t *Templates) parseFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	ext := filepath.Ext(f.Name())
	if f.IsDir() || !t.check(ext) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("templates: error opening path: %q %v", path, err)
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("templates: error reading file contents, path: %q %v", path, err)
	}

	subPath := strings.Replace(path, t.stripPrefix, "", 1)
	if strings.Contains(path, "/view/") || strings.Contains(path, "/views/") {
		t.AddView(subPath, string(contents))
	} else {
		t.AddPartial(subPath, string(contents))
	}

	return nil
}

func (t *Templates) check(ext string) bool {
	if len(t.Extensions) == 0 {
		return true
	}

	for x := range t.Extensions {
		if ext == x {
			return true
		}
	}

	return false
}

func (t *Templates) Parse() {
	if len(t.delimsLeft) == 0 {
		t.delimsLeft = "{{"
	}
	if len(t.delimsRight) == 0 {
		t.delimsRight = "}}"
	}

	// create a template that contains every partial
	tmpl := template.New("").Funcs(t.funcs).Delims(t.delimsLeft, t.delimsRight)
	for _, partial := range t.partials {
		tmpl = template.Must(tmpl.New(partial.key).Parse(partial.value))
	}

	// clone the main template to create the view template.
	// This enables the usage of Go's template "block" template feature
	for _, view := range t.templates {
		viewTmpl, _ := tmpl.Clone()
		viewTmpl = template.Must(viewTmpl.Parse(view.value))
		t.Templates[view.key] = viewTmpl
	}
}

func (t *Templates) Execute(w io.Writer, baseView, view string, data interface{}) error {
	setDefaultContentType(w)
	tmpl, ok := t.Templates[view]
	if !ok {
		return errTemplateNotFound{view}
	}

	if err := tmpl.ExecuteTemplate(w, baseView, data); err != nil {
		return fmt.Errorf("templates: ExecuteTemplate error, baseView %q view %q %v", baseView, view, err)
	}

	return nil
}

func (t *Templates) MustExecute(w io.Writer, baseView, view string, data interface{}) {
	if err := t.Execute(w, baseView, view, data); err != nil {
		panic(err.Error())
	}
}

func (t *Templates) ExecuteSingle(w io.Writer, view string, data interface{}) error {
	setDefaultContentType(w)

	tmpl, ok := t.Templates[view]
	if !ok {
		return errTemplateNotFound{view}
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("templates: Execute error, view %q %v", view, err)
	}

	return nil
}

func (t *Templates) MustExecuteSingle(w io.Writer, view string, data interface{}) {
	if err := t.ExecuteSingle(w, view, data); err != nil {
		panic(err)
	}
}

func (t *Templates) Render(baseView, view string, data interface{}) ([]byte, error) {
	buf := t.pool.get()
	defer func() {
		t.pool.put(buf)
	}()

	tmpl, err := t.get(view)
	if err != nil {
		return nil, fmt.Errorf("failed to get template in render func: %v", err)
	}
	if err := tmpl.ExecuteTemplate(buf, baseView, data); err != nil {
		return nil, fmt.Errorf("templates: ExecuteTemplate error, baseView %q view %q %v", baseView, view, err)
	}

	return buf.Bytes(), nil
}
func (t *Templates) RenderRequest(r *http.Request, baseView, view string, data interface{}) ([]byte, error) {
	buf := t.pool.get()
	defer func() {
		t.pool.put(buf)
	}()

	tmpl, err := t.get(view)
	if err != nil {
		return nil, fmt.Errorf("failed to get template in render request: %v", err)
	}
	tmpl.Funcs(funcmaps.RequestFuncMap(r))
	if err := tmpl.ExecuteTemplate(buf, baseView, data); err != nil {
		return nil, fmt.Errorf("templates: ExecuteTemplate error, baseView %q view %q %v", baseView, view, err)
	}

	return buf.Bytes(), nil
}
func (t *Templates) get(template string) (*template.Template, error) {
	tmp, ok := t.Templates[template]
	if !ok {
		return nil, errTemplateNotFound{template}
	}
	return tmp, nil
}
func (t *Templates) executeTemplate(tmp *template.Template, wr io.Writer, name string, data interface{}) error {
	return tmp.ExecuteTemplate(wr, name, data)
}
func (t *Templates) MustRender(baseView, view string, data interface{}) []byte {
	b, err := t.Render(baseView, view, data)
	if err != nil {
		panic(err)
	}
	return b
}

func (t *Templates) RenderSingle(view string, data interface{}) ([]byte, error) {
	buf := t.pool.get()
	defer func() {
		t.pool.put(buf)
	}()

	if err := t.ExecuteSingle(buf, view, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *Templates) MustRenderSingle(view string, data interface{}) []byte {
	b, err := t.RenderSingle(view, data)
	if err != nil {
		panic(err)
	}
	return b
}

func setDefaultContentType(w io.Writer) {
	if rw, ok := w.(http.ResponseWriter); ok {
		if len(rw.Header().Get("Content-Type")) == 0 {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
	}
}
