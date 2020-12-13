package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/joho/godotenv"
)

type Employee struct {
	Id   int
	Name string
	City string
}
type Ticket struct {
	ID          int
	User        string
	Description string
	Details     string
	CcList      string
}

const (
	mysqlHostEnvVarName     = "MYSQL_HOST"
	mysqlPortEnvVarName     = "MYSQL_PORT"
	mysqlUserEnvVarName     = "MYSQL_USER"
	mysqlPasswordEnvVarName = "MYSQL_PASSWORD"
	mysqlDbNameEnvVarName   = "MYSQL_DBNAME"
)

var db *sql.DB

// Simple helper function to read an environment or return a default value
// https://dev.to/craicoverflow/a-no-nonsense-guide-to-environment-variables-in-go-a2f
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func defaultDBConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbHost := getEnv(mysqlHostEnvVarName, "localhost")
	dbPort := getEnv(mysqlPortEnvVarName, "3306")
	dbUser := getEnv(mysqlUserEnvVarName, "admin")
	dbPass := getEnv(mysqlPasswordEnvVarName, "admin")
	dbName := getEnv(mysqlDbNameEnvVarName, "db")
	log.Printf("Using DB %s on %s:%s\n", dbDriver, dbHost, dbPort)

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}
func sqliteDBConn() (db *sql.DB) {
	sqliteFilePath := "/tmp/octicketing.db"
	log.Printf("Using local sqlite file at %s\n", sqliteFilePath)

	sqliteDB, err := sql.Open("sqlite3", sqliteFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	statement, err := sqliteDB.Prepare("CREATE TABLE IF NOT EXISTS tickets (  id integer  NOT NULL PRIMARY KEY AUTOINCREMENT,  user varchar(8) NOT NULL,  description varchar(30) NOT NULL,  details varchar(300) NOT NULL,  cc_list varchar(72));")
	if err != nil {
		fmt.Println(err.Error())
	}
	statement.Exec()

	return sqliteDB
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func handleDbError(wr io.Writer, e error) {
	log.Println("handling db error. \n" + e.Error())
	tmpl.ExecuteTemplate(wr, "NoDBConn", nil)

}
func Index(w http.ResponseWriter, r *http.Request) {
	selDB, err := db.Query("SELECT id, user, description, details, COALESCE(cc_list, '') FROM tickets ORDER BY id DESC")
	if err != nil {

		handleDbError(w, err)
		return
	}
	defer selDB.Close()
	ticketItem := Ticket{}
	ticketList := []Ticket{}

	for selDB.Next() {
		var id int
		var user, description, details, ccList string
		err = selDB.Scan(&id, &user, &description, &details, &ccList)
		if err != nil {
			handleDbError(w, err)
			return
		}
		ticketItem.ID = id
		ticketItem.User = user
		ticketItem.Description = description
		ticketItem.Details = details
		ticketItem.CcList = ccList
		ticketList = append(ticketList, ticketItem)
	}

	tmpl.ExecuteTemplate(w, "Index", ticketList)

}

func Show(w http.ResponseWriter, r *http.Request) {
	nID := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT id, user, description, details, COALESCE(cc_list, '') FROM tickets WHERE id=?", nID)
	if err != nil {
		handleDbError(w, err)
		return
	}
	defer selDB.Close()
	ticketItem := Ticket{}

	for selDB.Next() {
		var id int
		var user, description, details, ccList string

		err = selDB.Scan(&id, &user, &description, &details, &ccList)
		if err != nil {
			handleDbError(w, err)
			return
		}
		ticketItem.ID = id
		ticketItem.User = user
		ticketItem.Description = description
		ticketItem.Details = details
		ticketItem.CcList = ccList

	}
	tmpl.ExecuteTemplate(w, "Show", ticketItem)

}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	nID := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT id, user, description, details, COALESCE(cc_list, '') FROM tickets WHERE id=?", nID)
	if err != nil {
		handleDbError(w, err)
		return
	}
	defer selDB.Close()
	ticketItem := Ticket{}

	for selDB.Next() {
		var id int
		var user, description, details, ccList string
		err = selDB.Scan(&id, &user, &description, &details, &ccList)
		if err != nil {
			handleDbError(w, err)
			return
		}
		ticketItem.ID = id
		ticketItem.User = user
		ticketItem.Description = description
		ticketItem.Details = details
		ticketItem.CcList = ccList
	}
	tmpl.ExecuteTemplate(w, "Edit", ticketItem)

}

func Insert(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := r.FormValue("user")
		description := r.FormValue("description")
		details := r.FormValue("details")
		ccList := r.FormValue("ccList")
		insForm, err := db.Prepare("INSERT INTO tickets(user, description, details, cc_list) VALUES(?,?,?,?)")
		if err != nil {
			handleDbError(w, err)
			return
		}
		defer insForm.Close()
		insForm.Exec(user, description, details, ccList)
		log.Println("INSERT: user: " + user + " | description: " + description)
	}

	http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := r.FormValue("user")
		description := r.FormValue("description")
		details := r.FormValue("details")
		ccList := r.FormValue("ccList")
		uid := r.FormValue("uid")
		insForm, err := db.Prepare("UPDATE tickets SET user=?, description=?, details=?, cc_list=? where id=?")
		if err != nil {
			handleDbError(w, err)
			return
		}
		defer insForm.Close()
		insForm.Exec(user, description, details, ccList, uid)
		log.Println("INSERT: user: " + user + " | description: " + description)
	}

	http.Redirect(w, r, "/", 301)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	log.Println("in delete")
	ticketID := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM tickets WHERE id=?")
	if err != nil {
		handleDbError(w, err)
		return
	}
	defer delForm.Close()

	delForm.Exec(ticketID)
	log.Println("DELETE")

	http.Redirect(w, r, "/", 301)
}

func Health(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		w.Write([]byte("healthy"))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func setLocalDB() {
	if value, exists := os.LookupEnv("DB_ENGINE"); exists {
		if value == "SQLITE" {
			db = sqliteDBConn()
		}
	}

}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if value, exists := os.LookupEnv("DB_ENGINE"); exists {
		if value == "SQLITE" {
			db = sqliteDBConn()
		}
	} else {

		db = defaultDBConn()
	}

}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	log.Println("Server started on: http://localhost:8080")

	r.Get("/", Index)
	r.Get("/show", Show)
	r.Get("/new", New)
	r.Get("/edit", Edit)
	r.Post("/insert", Insert)
	r.Post("/update", Update)
	r.Get("/delete", Delete)
	r.Get("/health", Health)
	log.Fatal(http.ListenAndServe(":8080", r))

}
