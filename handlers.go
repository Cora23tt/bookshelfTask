package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"bookshelf/database"
	"bookshelf/models"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func createUser(c *gin.Context) {
	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if user.Email == "" || user.Name == "" || user.Key == "" || user.Secret == "" {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	newuser, err := database.CreateUser(user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unnable create new user. ERROR: " + err.Error()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unnable create new user. ERROR: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": newuser, "isOk": true, "message": "ok"})
}

func getUserInfo(c *gin.Context) {

	if !CompareSign(c, "") {
		return
	}

	Key := c.GetHeader("Key")
	user, err := database.GetUserByKey(Key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user, "isOk": true, "message": "ok"})
}

func searchBooks(c *gin.Context) {

	if !CompareSign(c, "") {
		return
	}

	books, err := database.SearchBooksByTitle(c.GetHeader("Key"), c.Param("title"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unnable search book. ERROR: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": books, "isOk": true, "message": "ok"})
}

func createBook(c *gin.Context) {

	// getting request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// checking sing
	if !CompareSign(c, string(body)) {
		return
	}

	// getting book isbn from reqBody
	var isbn models.ISBN
	if err = json.Unmarshal(body, &isbn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// getting book data from openlibrary by isbn ======================
	book := GetBookOpenLib(isbn)

	// creating book
	Book, err := database.CreateBook(book, c.GetHeader("Key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.BookResponse{IsOK: true, Message: "ok", Data: models.BookStatus{Status: 0, Book: Book}})
}

func getAllBooks(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	var isbn models.ISBN
	json.Unmarshal(body, &isbn)

	if !CompareSign(c, "") {
		return
	}

	books, err := database.GetAllBooks(c.GetHeader("Key"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.BooksResponse{Data: books, IsOK: true, Message: "ok"})
}

func editBook(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	var Status models.Status
	json.Unmarshal(body, &Status)

	if !CompareSign(c, string(body)) {
		return
	}

	// getting param book_id
	s := c.Param("id")
	book_id, err := strconv.Atoi(s)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	// Edit book status
	book, err := database.EditStatus(c.GetHeader("Key"), book_id, Status.Status)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "isOk": true, "message": "unable to edit"})
		return
	}
	c.JSON(http.StatusOK, models.BookResponse{Data: models.BookStatus{Book: book, Status: Status.Status}, IsOK: true, Message: "ok"})
}

func deleteBook(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	var Status models.Status
	json.Unmarshal(body, &Status)

	if !CompareSign(c, string(body)) {
		return
	}

	// getting param book_id
	s := c.Param("id")
	book_id, err := strconv.Atoi(s)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "isOk": true, "message": "unable to authorize"})
		return
	}

	// user, err := database.GetUserByKey(c.GetHeader("Key"))
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
	// 	return
	// }

	// Delete book
	books, err := database.DeleteBook(c.GetHeader("Key"), book_id)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return
	}

	c.JSON(http.StatusOK, models.BooksResponse{Data: books, IsOK: true, Message: "ok"})
}

func GetBookOpenLib(isbn models.ISBN) models.Book {

	requestURL := fmt.Sprintf("https://openlibrary.org/%s%v.json", "isbn/", isbn.ISBN)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	body, _ := io.ReadAll(res.Body)

	var PublishDate models.BookPublishDate
	var Authors models.Authors
	var Title models.BookTtitle
	var ISBN10 models.ISBN10
	var ISBN13 models.ISBN13
	var BookCovers models.BookCovers
	var NumberOfPages models.NumberOfPages

	json.Unmarshal(body, &PublishDate)
	json.Unmarshal(body, &Authors)
	json.Unmarshal(body, &Title)
	json.Unmarshal(body, &ISBN10)
	json.Unmarshal(body, &ISBN13)
	json.Unmarshal(body, &BookCovers)
	json.Unmarshal(body, &NumberOfPages)

	var _isbn string
	if ISBN13.ISBN13 != "" {
		_isbn = ISBN13.ISBN13
	} else if ISBN10.ISBN10 != "" {
		_isbn = ISBN13.ISBN13
	} else {
		_isbn = isbn.ISBN
	}

	_coverURL := ""
	if len(BookCovers.Covers) != 0 {
		_coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%v-L.jpg", BookCovers.Covers[0])
	}

	// getting author name from openlibrary by link
	var Author models.Author
	_author := ""
	if len(Authors.Authors) != 0 {
		authorURL := Authors.Authors[0].URL
		_requestURL := fmt.Sprintf("https://openlibrary.org%s.json", authorURL)
		_res, err := http.Get(_requestURL)
		if err != nil {
			log.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}
		_body, _ := io.ReadAll(_res.Body)
		json.Unmarshal(_body, &Author)
	}
	if Author.Name != "" {
		_author = Author.Name
	}

	// convert str publish_date to integer, if that possible
	_published := 0
	if PublishDate.PublishDate != "" {

		re := regexp.MustCompile("[0-9]+")
		_dateStr := re.FindAllString(PublishDate.PublishDate, -1)

		if len(_dateStr) > 0 {
			_published, _ = strconv.Atoi(_dateStr[0])
			for _, v := range _dateStr {
				_v, err := strconv.Atoi(v)
				if err != nil {
					break
				}
				if _published < _v {
					_published = _v
				}
			}
		}
	}

	_pages := 0
	if NumberOfPages.Pages != 0 {
		_pages = NumberOfPages.Pages
	}

	return models.Book{ID: 0, ISBN: _isbn, Title: Title.Title, Cover: _coverURL, Author: _author, Published: _published, Pages: _pages}
}

func CompareSign(c *gin.Context, body string) bool {
	Sign := c.GetHeader("Sign")
	Key := c.GetHeader("Key")

	user, err := database.GetUserByKey(Key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return false
	}

	// generating sign
	hasher := md5.New()
	hasher.Write([]byte(c.Request.Method + "http://" + HOST + PORT + c.Request.RequestURI + body + user.Secret))
	generatedSign := hex.EncodeToString(hasher.Sum(nil))

	// compare sign and generated
	if generatedSign != Sign {
		c.JSON(http.StatusUnauthorized, gin.H{"data": "the sign is invalid", "isOk": true, "message": "unable to authorize"})
		return false
	}

	return true
}
