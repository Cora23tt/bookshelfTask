package database

import (
	"bookshelf/models"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Dbinstance struct {
	DB *sqlx.DB
}

var Postgres Dbinstance

func ConnectDB() {
	db, err := sqlx.Connect("postgres", "user=postgres password=3729 dbname=bookshelf sslmode=disable") //
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}

	log.Println("DB connect success")

	Postgres = Dbinstance{
		DB: db,
	}
}

func CreateUser(user models.User) (newUser models.User, err error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO users (name, email, key, secret) VALUES ('%v', '%v', '%v', '%v') RETURNING id`,
		user.Name, user.Email, user.Key, user.Secret)
	if err := Postgres.DB.QueryRow(query).Scan(&id); err != nil {
		return models.User{}, err
	}

	return models.User{ID: id, Name: user.Name, Email: user.Email, Key: user.Key, Secret: user.Secret}, nil
}

func GetUserByKey(headerKey string) (user models.User, err error) {
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM users WHERE key='%v' LIMIT 1`, headerKey))
	if err != nil {
		return models.User{}, err
	}

	ok := rows.Next()
	if ok {
		if err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Key, &user.Secret); err != nil {
			return models.User{}, err
		}
	}
	return user, nil
}

func SearchBooksByTitle(key, title string) (arr []models.Book, err error) {

	rows, err := Postgres.DB.Query(fmt.Sprintf(`
	SELECT books_tbl.id, books_tbl.isbn, books_tbl.title, books_tbl.cover, books_tbl.author, books_tbl.published, books_tbl.pages 
	FROM books books_tbl
	INNER JOIN user_books users_tbl
	ON books_tbl.id = users_tbl.book_id
	WHERE users_tbl.user_id=(SELECT id FROM users WHERE key='%s') 
	AND POSITION('%s' in books_tbl.title)>0`, key, title))
	if err != nil {
		return nil, err
	}

	i := 0
	var book models.Book
	for rows.Next() {
		if err = rows.Scan(&book.ID, &book.ISBN, &book.Title, &book.Cover, &book.Author, &book.Published, &book.Pages); err != nil {
			return nil, err
		}
		arr = append(arr, book)
		i++
	}
	return arr, nil
}

func CreateBook(book models.Book, Key string) (Book models.Book, err error) {

	// create book if not exists
	query := `
	INSERT INTO books (isbn, title, cover, author, published, pages) 
	SELECT '%s' AS isbn, '%s' AS title, '%s' AS cover, '%s' AS author, %d AS published, %d AS pages
	WHERE NOT EXISTS (SELECT id FROM books WHERE isbn='%s');
	`
	query = fmt.Sprintf(query, book.ISBN, book.Title, book.Cover, book.Author, book.Published, book.Pages, book.ISBN)
	if _, err := Postgres.DB.Exec(query); err != nil {
		return models.Book{}, err
	}

	// bind book to user
	query = `
	INSERT INTO user_books 
	(user_id, book_id, status) 
	SELECT (SELECT id FROM users WHERE key='%s') AS user_id, (SELECT id FROM books WHERE isbn='%s') AS book_id, 0 as status
	WHERE NOT EXISTS 
	(SELECT user_id FROM user_books WHERE user_id=(SELECT id FROM users WHERE key='%s') AND book_id=(SELECT id FROM books WHERE isbn='%s'))
	RETURNING book_id;
	`
	var id int
	query = fmt.Sprintf(query, Key, book.ISBN, Key, book.ISBN)
	if err := Postgres.DB.QueryRow(query).Scan(&id); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return models.Book{}, errors.New("already exists")
		}
		return models.Book{}, err
	}
	return models.Book{ID: id, ISBN: book.ISBN, Title: book.Title, Cover: book.Cover, Author: book.Author, Published: book.Published, Pages: book.Pages}, nil
}

func GetAllBooks(key string) (books []models.BookStatus, err error) {
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT id, isbn, title, cover, author, published, pages, status FROM books INNER JOIN user_books ON books.id=user_books.book_id WHERE user_id=(SELECT id FROM users WHERE key='%s')`, key))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var book models.BookStatus
		if err = rows.Scan(&book.Book.ID, &book.Book.ISBN, &book.Book.Title, &book.Book.Cover, &book.Book.Author, &book.Book.Published, &book.Book.Pages, &book.Status); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

func EditStatus(key string, book_id int, Status int) (Book models.Book, err error) {

	// update book status for user
	res, err := Postgres.DB.Exec(`UPDATE user_books SET status=$1 WHERE 
	user_id=(SELECT id FROM users WHERE key=$2) AND book_id=$3`, Status, key, book_id)
	if err != nil {
		return models.Book{}, err
	}

	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		log.Println("there!")
		return models.Book{}, fmt.Errorf("book with id %d is not exist", book_id)
	}

	// find book
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM books WHERE id='%d'`, book_id))
	if err != nil {
		return models.Book{}, err
	}

	if rows.Next() {
		if err = rows.Scan(&Book.ID, &Book.ISBN, &Book.Title, &Book.Cover, &Book.Author, &Book.Published, &Book.Pages); err != nil {
			return models.Book{}, err
		}
	}

	return Book, nil
}

func DeleteBook(key string, book_id int) (books []models.BookStatus, err error) {

	// delete book from user books
	if _, err := Postgres.DB.Exec(`DELETE FROM user_books WHERE user_id=(SELECT id FROM users WHERE key=$1) and book_id=$2`, key, book_id); err != nil {
		return nil, err
	}

	// check, if book useless, then delete
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM user_books where book_id=%d`, book_id))
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		if _, err := Postgres.DB.Exec(`DELETE FROM books WHERE id=$1`, book_id); err != nil {
			return nil, err
		}
	}

	return GetAllBooks(key)
}
