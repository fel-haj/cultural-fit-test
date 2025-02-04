package handler

import (
  "net/http"
  "html/template"
  "fmt"
)

type PageData struct {
  Title string
}

func IndexHandler() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
      data := PageData {
        Title: "Mentoref - Personality Test",
      }

      index := template.Must(template.ParseFiles("web/templates/index.html"))
      
      err := index.Execute(w, data)
      if err != nil {
        http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
        return
      }
    }
  }
}
