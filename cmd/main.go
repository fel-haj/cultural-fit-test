package main

import (
  "net/http"
  "mentoref/db"
  "mentoref/internal/handler"
  "log"
)

func main() {
  mux := http.NewServeMux()
  dbClient := db.ConnectDB()
  defer dbClient.Close()

  fs := http.FileServer(http.Dir("./web/css"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

  mux.HandleFunc("/", handler.IndexHandler())
  mux.HandleFunc("/question", handler.QuestionHandler(dbClient))

  log.Fatal(http.ListenAndServe(":3000", mux))
}
