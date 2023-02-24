package main

import (
	"bookshelf/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	PORT = ":9090"
	HOST = "127.0.0.1"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies([]string{HOST})
	database.ConnectDB()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	r.POST("/signup", createUser)
	r.GET("/myself", getUserInfo)
	r.GET("/books/:title", searchBooks)
	r.POST("/books", createBook)
	r.GET("/books", getAllBooks)
	r.PATCH("/books/:id", editBook)
	r.DELETE("/books/:id", deleteBook)

	r.Run(PORT)

}
