package tmplts

import "html/template"

func (t *Templates) AddView(name string, tmpl string) {
	t.templates = append(t.templates, keyValue{
		key:   name,
		value: tmpl,
	})
}

func (t *Templates) AddPartial(name string, tmpl string) {
	t.partials = append(t.partials, keyValue{
		key:   name,
		value: tmpl,
	})
}
func (t *Templates) AddFunc(name string, f interface{}) {
	t.funcs[name] = f
}

func (t *Templates) AddFuncs(funcMaps ...template.FuncMap) {
	for _, funcs := range funcMaps {
		for k, v := range funcs {
			t.funcs[k] = v
		}
	}
}

func (t *Templates) UseExts(extensions []string) {
	exts := make(map[string]bool)
	for _, ext := range extensions {
		exts[ext] = true
	}
	t.Extensions = exts
}
