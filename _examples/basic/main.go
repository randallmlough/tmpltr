package main

import (
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/go-chi/chi"
	"github.com/randallmlough/tmplts/funcmaps"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/randallmlough/tmplts"
)

var (
	// templates global that will contain all of our parsed temlates from the templates directory
	tmpls *tmplts.Templates
	port  = ":8083"
	host  = "http://localhost" + port
)

var (
	css     = []string{fmt.Sprintf("%v/static/css/main.css", host)}
	scripts = []string{fmt.Sprintf("%v/static/js/main.js", host)}
)

type User struct {
	ID   int
	Name string
	URL  string
}

var (
	users = []User{
		{
			ID:   1,
			Name: "John Doe",
		},
		{
			ID:   2,
			Name: "Jane Doe",
		},
	}
)

// parse the templates in the template directory
func init() {
	var err error
	tmpls, err = tmplts.New().ParseDir("./templates", "templates/")
	if err != nil {
		log.Fatal(err)
	}
	tmpls.AddRequestFuncs(funcmaps.RequestFuncMap)
	tmpls.AddFuncs(
		sprig.FuncMap(),
	)
}

func main() {
	tmpls.Parse()

	r := chi.NewMux()
	// web pages
	r.Get("/", renderPage("index", "Index Page Title"))
	r.Get("/about", renderPage("about", "About Page Title"))
	r.Get("/contact", renderPage("contact/single", "Contact us"))
	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		view := "users/list"
		b, err := tmpls.RenderRequest(r, "base.html", "views/"+view+".html", map[string]interface{}{
			"Title":   "Users Page",
			"Css":     css,
			"Scripts": scripts,
			"Menu":    activeNav(view),
			"Users":   users,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(b)
	})
	r.Get("/user/{userID}", func(w http.ResponseWriter, r *http.Request) {
		view := "users/single"
		tmp := chi.URLParam(r, "userID")
		userID, err := strconv.Atoi(tmp)
		if err != nil {
			http.Error(w, "error", 500)
		}
		// get the user from ID
		user := User{}
		for _, u := range users {
			if u.ID == userID {
				user = u
			}
		}

		// for simplicity, lets assume this means we didn't find a user
		if user.Name == "" {
			http.Error(w, "no user found", 404)
		}
		b, err := tmpls.RenderRequest(r, "base.html", "views/"+view+".html", map[string]interface{}{
			"Title":   user.Name,
			"Css":     css,
			"Scripts": scripts,
			"Menu":    activeNav(view),
			"User":    user,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(b)
	})

	// serve static files
	FileServer(r, "/static", http.Dir("./static/"))

	// Start http server
	log.Println("Server stared on " + host)
	log.Fatal(http.ListenAndServe(port, r))
}

func renderPage(view string, title string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := tmpls.RenderRequest(r, "base.html", "views/"+view+".html", map[string]interface{}{
			"Title":   title,
			"Css":     css,
			"Scripts": scripts,
			"Menu":    activeNav(view),
			"Name":    "John",
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(b)
	}
}

type navItem struct {
	Name  string
	Attrs map[template.HTMLAttr]string
}

func activeNav(active string) []navItem {
	// create menu items
	about := navItem{
		Name: "About",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/about",
			"title": "About Page",
		},
	}
	home := navItem{
		Name: "Home",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/",
			"title": "Home Page",
		},
	}
	contact := navItem{
		Name: "Contact",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/contact",
			"title": "Contact Page",
		},
	}
	users := navItem{
		Name: "Users",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/users",
			"title": "Users",
		},
	}
	// set active menu class
	switch active {
	case "about":
		about.Attrs["class"] = "active"
	case "home":
		home.Attrs["class"] = "active"
	}

	return []navItem{home, about, contact, users}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
