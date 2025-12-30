package handler

import (
	"embed"
	"errors"
	"html/template"
	"net/http"
)

//go:embed templates/index.html
var homeTemplateFS embed.FS

type homeData struct {
	BaseURL string
	Subject string
	Status  string
	Color   string
	Style   string
}

func parseHomeTemplate() (*template.Template, error) {
	tpl, err := template.ParseFS(homeTemplateFS, "templates/index.html")
	if err != nil {
		return nil, err
	}
	if tpl == nil {
		return nil, errors.New("home template is nil")
	}
	return tpl, nil
}

// Home renders the UI landing page.
func (h *Handler) Home(w http.ResponseWriter, req *http.Request) {
	scheme := "http"
	if forwarded := req.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	} else if req.TLS != nil {
		scheme = "https"
	}
	host := req.Host
	if host == "" {
		host = "localhost"
	}
	data := homeData{
		BaseURL: scheme + "://" + host,
		Subject: "build",
		Status:  "passing",
		Color:   "green",
		Style:   "flat",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.home.Execute(w, data); err != nil {
		writeError(w, http.StatusInternalServerError, "render template")
	}
}
