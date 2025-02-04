package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	DBURL     string
}

type User struct {
  ID uint32
  Email string
  CurrQuestNo uint32
}

type Question struct {
	QuestionNo uint32
  Question string
	ChoiceA string
	ChoiceB string
	ChoiceC string
	ChoiceD string
	ConnectA string
	ConnectB string
	ConnectC string
	ConnectD string
  ConnectTheme string
  Chapter string
}

func ConnectDB() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to database")

	return db
}
