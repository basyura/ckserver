package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Response struct {
	Method      string
	Path        string
	QueryParams []KeyValue
	FormParams  []KeyValue
	Params      []KeyValue
}

type KeyValue struct {
	Key   string
	Value string
}

// curl --noproxy "*" "http://localhost:2345/result?a=b&c=d"
// curl --noproxy "*" -F "a=b" -F "c=d" -X POST http://localhost:2345/result
func main() {
	var exePath string
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	execDir := filepath.Dir(exePath)

	e := echo.New()
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob(filepath.Join(execDir, "view", "*.html"))),
	}

	e.Static("/", filepath.Join(execDir, "view/index.html"))

	e.GET("/result", func(c echo.Context) error {
		return c.Render(http.StatusOK, "result", parse(c, "GET"))
	})
	e.POST("/result", func(c echo.Context) error {
		return c.Render(http.StatusOK, "result", parse(c, "POST"))
	})

	go e.Logger.Fatal(e.Start(":2345"))
}

func parse(c echo.Context, method string) *Response {
	res := &Response{
		Method:      method,
		QueryParams: []KeyValue{},
		FormParams:  []KeyValue{},
		Params:      []KeyValue{},
	}
	res.Path = c.Path()
	for k, v := range c.QueryParams() {
		res.QueryParams = append(res.QueryParams, KeyValue{Key: k, Value: strings.Join(v, ",")})
	}

	if fp, err := c.FormParams(); err != nil {
		res.FormParams = append(res.FormParams, KeyValue{Key: "error", Value: err.Error()})
	} else {
		for k, v := range fp {
			res.FormParams = append(res.FormParams, KeyValue{Key: k, Value: strings.Join(v, ",")})
		}
	}

	for _, k := range c.ParamNames() {
		res.Params = append(res.Params, KeyValue{Key: k, Value: c.Param(k)})
	}

	return res
}
