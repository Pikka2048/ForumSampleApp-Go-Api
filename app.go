package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type PostStruct struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

var db *sql.DB

const (
	MaxBodyBytes = 1024 * 8
)

func SqliteInit() {
	os.Remove("./mysqlite.db")
	var err error
	db, err = sql.Open("sqlite3", "./mysqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	//Make Table
	sql := "create table post (id integer not null primary key,name text,content text);"

	_, err = db.Exec(sql)
	if err != nil {
		log.Printf("%q: %s\n", err, sql)
		return
	}
}

func InsertPostDataSql(data PostStruct) {
	var err error
	sql := "insert into post values($1,$2,$3);"
	_, err = db.Exec(sql, data.Id, data.Name, data.Content)
	if err != nil {
		log.Printf("%q: %s\n", err, sql)
		return
	}
}
func StoreSql2Sturct() []PostStruct {
	var err error
	//HACK Limit Required!
	sql := "select * from post"

	rows, err := db.Query(sql)
	if err != nil {
		log.Printf("%q: %s\n", err, sql)
		return nil
	}
	defer rows.Close()
	// Store Sql in Struct
	var Results []PostStruct
	for rows.Next() {
		var id int
		var name string
		var content string
		var result PostStruct
		err := rows.Scan(&id, &name, &content)
		if err != nil {
			log.Printf("%q: %s\n", err, sql)
		}
		result.Id = id
		result.Name = name
		result.Content = content
		Results = append(Results, result)
	}
	return Results
}
func bodySizeMiddleware(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	c.Request.Body = http.MaxBytesReader(w, c.Request.Body, MaxBodyBytes)

	c.Next()
}

func main() {
	SqliteInit()

	r := gin.Default()
	r.Use(bodySizeMiddleware)
	// CORS SETUP
	r.Use(cors.New(cors.Config{
		// Allow Access IP
		AllowOrigins: []string{
			"http://localhost:8081",
		},
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		AllowCredentials: true,
		//Cache
		MaxAge: 24 * time.Hour,
	}))

	// ROUTE
	r.POST("/api/post", Post)
	r.Run()
	// END
	defer db.Close()
}

func Post(c *gin.Context) {
	var json PostStruct
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//No Names ALLOWED
	if json.Name == "" {
		json.Name = "anonymous"
	}
	if json.Content != "" {
		InsertPostDataSql(json)
	}
	resPostStruct := StoreSql2Sturct()
	c.JSON(http.StatusOK, resPostStruct)
}
