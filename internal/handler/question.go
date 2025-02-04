package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"mentoref/db"
)

type Choices struct {
  ID string
  Prompt string
}

type QuestionData struct {
	UserID     uint32
	QuestionNo uint32
	Question   string
  Choices    []Choices
	Chapter    string
	Options    []uint8
}

type MessageData struct {
  Connection string
  ConnectionTheme string
  Chapter string
  QuestionNo string
  End bool
}

func ShowEndMessage(w http.ResponseWriter) {
  message := template.Must(template.ParseFiles("web/templates/message.html"))
  msg := MessageData {
    End: true,
  }
  err := message.Execute(w, msg)
  if err != nil {
    http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
  }
}

func QuestionHandler(dbClient *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.URL.Query().Has("end") {
        ShowEndMessage(w) 
        return
      }
			var email string
			if r.URL.Query().Has("register") {
				email = r.FormValue("email")
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    email,
					HttpOnly: true,
					Expires:  time.Now().Add(120 * time.Minute),
				})
			}

			cookie, err := r.Cookie("session")
			if err == nil {
				email = cookie.Value
			}

      var user db.User
      err = dbClient.QueryRow(`SELECT id, currquestno FROM candidates WHERE email = $1`, email).Scan(&user.ID, &user.CurrQuestNo)
      if err == sql.ErrNoRows {
        _, err = dbClient.Exec(`INSERT INTO candidates (email, currquestno) VALUES ($1, $2)`, email, 0)
        if err != nil {
          log.Printf("Database error during user insertion: %v", err)
          return
        }
        err = dbClient.QueryRow(`SELECT id FROM candidates WHERE email = $1`, email).Scan(&user.ID)
        if err != nil {
          log.Printf("Database error: %v", err)
          w.WriteHeader(http.StatusBadRequest)
          w.Write([]byte(`<div>An internal error has occurred.<br>Please try again later.</div>`))
          return
        }
      }

      if user.CurrQuestNo == 40 {
        ShowEndMessage(w)
        return
      }

			var modifier uint32
			if r.URL.Query().Has("quest_no") {
				questNo, _ := strconv.ParseUint(r.URL.Query().Get("quest_no"), 10, 32)
				modifier = uint32(questNo)
			} else {
				modifier = user.CurrQuestNo + 1
			}

			var question db.Question
			err = dbClient.QueryRow(`SELECT questionno, question, choicea, choiceb, choicec, choiced, chapter FROM questions WHERE questionno = $1;`, modifier).Scan(&question.QuestionNo, &question.Question, &question.ChoiceA, &question.ChoiceB, &question.ChoiceC, &question.ChoiceD, &question.Chapter)
			if err == sql.ErrNoRows {
				log.Printf("No rows found: %v", err)
			} else if err != nil {
				log.Printf("Database error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`<div>An internal error has occurred.<br>Please try again later.</div>`))
			}

      choices := []Choices {
				{ID: "a", Prompt: question.ChoiceA},
				{ID: "b", Prompt: question.ChoiceB},
				{ID: "c", Prompt: question.ChoiceC},
				{ID: "d", Prompt: question.ChoiceD},
			}

			data := QuestionData {
				UserID:     user.ID,
				QuestionNo: question.QuestionNo,
				Question:   question.Question,
        Choices:    choices,
				Chapter:    question.Chapter,
				Options:    []uint8{1, 2, 3, 4, 5},
			}

			questionTmp := template.Must(template.ParseFiles("web/templates/question.html"))
			err = questionTmp.Execute(w, data)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
				return
			}
		}

		if r.Method == "POST" {
			userid := r.FormValue("userid")
			questionno := r.FormValue("questionno")

			choices := map[string]string{
				"a": r.FormValue("a"),
				"b": r.FormValue("b"),
				"c": r.FormValue("c"),
				"d": r.FormValue("d"),
			}

      fmt.Println(choices)

			seen := make(map[string]bool)
			for _, v := range choices {
				if seen[v] {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`Each choice must have a unique value.`))
					return
				}
        seen[v] = true
			}

      conn := ""
			for k, v := range choices {
				if v == "1" {
          conn = fmt.Sprintf("connection%s", k)
				}
			}
			if conn == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`One choice must be set to 1.`))
				return
			}

			_, err := dbClient.Exec(`INSERT INTO choices (userid, questionno, choicea, choiceb, choicec, choiced) VALUES ($1, $2, $3, $4, $5, $6)`, userid, questionno, choices["a"], choices["b"], choices["c"], choices["d"])
			if err != nil {
				log.Printf("Database error: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
			}

      _, err = dbClient.Exec(`UPDATE candidates SET currquestno = ($1) WHERE id = ($2)`, questionno, userid)
			if err != nil {
				log.Printf("Database error: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
			}

      query := fmt.Sprintf("SELECT %s, connectiontheme, chapter FROM questions WHERE questionno = %s;", conn, questionno)
      var msg MessageData
      err = dbClient.QueryRow(query).Scan(&msg.Connection, &msg.ConnectionTheme, &msg.Chapter)
      msg.QuestionNo = questionno

      message := template.Must(template.ParseFiles("web/templates/message.html"))
      err = message.Execute(w, msg)
      if err != nil {
        http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
        return
      }
		}
	}
}
