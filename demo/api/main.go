package main

import (
	"fmt"
	"time"
	// "html"
	// "log"
	// "net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	name  string
	email string
}

// CURD
func test(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "ping !",
	})
}
func GetAllUser(ctx *gin.Context) {
	db := sqlConnect()
	defer db.Close()

	var users []User
	all, err := db.Query("SELECT * FROM Users")
	if err != nil {
		fmt.Println("Error when get all Users")
		fmt.Println(err)
		return
	}
	for all.Next() {
		var temp User
		all.Scan(&temp.name, &temp.email)
		users = append(users, temp)
	}
	fmt.Println(users)

	ctx.JSON(200, gin.H{"name": users[0].name, "email": users[0].email})
	ctx.JSON(200, gin.H{"name": users[0].name, "email": users[0].email})
}

func main() {
	db := sqlConnect()
	_, err := db.Query("CREATE TABLE Users (name VARCHAR(255), email VARCHAR(255))")

	if err != nil {
		fmt.Println("Error Query ------------:\n ", err)
	}

	db.Query("INSERT INTO Users VALUES ( 'minh', 'idol' )")
	db.Query("INSERT INTO Users VALUES ( 'an', 'ga bap' )")
	db.Close()

	router := gin.Default()

	router.GET("/test", test)
	router.GET("/", GetAllUser)

	router.Run(":8081")

}

func sqlConnect() *sql.DB {
	fmt.Println("Connect MySQL !!")

	DBMS := "mysql"
	USER := "go_test"
	PASS := "password"
	PROTOCOL := "tcp(db:3306)"
	DBNAME := "go_database"

	CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME

	db, err := sql.Open(DBMS, CONNECT)
	if err != nil {
		fmt.Println("\n--------------- ERROR OPEN -------------\n")
		panic(err.Error())
	}

	err = db.Ping()
	// if there is an error opening the connection, handle it
	for err != nil {
		fmt.Println("\n--------------- ERROR PING -------------\n")
		time.Sleep(500 * time.Millisecond)
		err = db.Ping()
		// panic(err.Error())
	}
	fmt.Println("\n------------------ PING ----------------\n")

	return db
	// defer db.Close()
}
