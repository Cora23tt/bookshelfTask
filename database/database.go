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

func SearchBooksByTitle(title string) (arr []models.Book, err error) {

	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM books WHERE POSITION('%s' IN title) > 0`, title))
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

func CreateBook(book models.Book, user_id int) (Book models.Book, err error) {

	// check if book alredy created
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM books WHERE title='%v' LIMIT 1`, book.Title))
	if err != nil {
		return models.Book{}, err
	}

	if rows.Next() {
		// scan finded book
		if err = rows.Scan(&Book.ID, &Book.ISBN, &Book.Title, &Book.Cover, &Book.Author, &Book.Published, &Book.Pages); err != nil {
			return models.Book{}, err
		}

		// check if bind is exist
		rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT status FROM user_books WHERE user_id=%v AND book_id=%v`, user_id, Book.ID))
		if err != nil {
			return models.Book{}, err
		}
		if rows.Next() {
			return models.Book{}, errors.New("alredy exist book")
		}

		// else bind user and book
		query := fmt.Sprintf(`INSERT INTO user_books VALUES ('%v', '%v', '%v')`, user_id, Book.ID, 0)
		_, err = Postgres.DB.Exec(query)
		if err != nil {
			return models.Book{}, err
		}
		return Book, nil
	}

	// if book not exist, insert this
	query := fmt.Sprintf(`INSERT INTO books (isbn, title, cover, author, published, pages) VALUES ('%v', '%v', '%v', '%v', '%d', %d) RETURNING id, isbn, title, cover, author, published, pages`,
		book.ISBN, book.Title, book.Cover, book.Author, book.Published, book.Pages)

	if err := Postgres.DB.QueryRow(query).Scan(&Book.ID, &Book.ISBN, &Book.Title, &Book.Cover, &Book.Author, &Book.Published, &Book.Pages); err != nil {
		return models.Book{}, err
	}
	// and bind to user
	query = fmt.Sprintf(`INSERT INTO user_books VALUES ('%v', '%v', '%v')`, user_id, Book.ID, 0)
	_, err = Postgres.DB.Exec(query)
	if err != nil {
		return models.Book{}, err
	}

	return Book, nil
}

func GetAllBooks(user_id int) (books []models.BookStatus, err error) {
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT id, isbn, title, cover, author, published, pages, status FROM books INNER JOIN user_books ON books.id=user_books.book_id WHERE user_id=%d`, user_id))
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

func EditStatus(user_id, book_id, Status int) (Book models.Book, err error) {

	// update book status for user
	if _, err := Postgres.DB.Exec(`UPDATE user_books SET status=$1 WHERE user_id=$2 AND book_id=$3`, Status, user_id, book_id); err != nil {
		return models.Book{}, err
	}

	// find book
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM books WHERE id='%d' LIMIT 1`, book_id))
	if err != nil {
		return models.Book{}, err
	}
	if rows.Next() {
		if err = rows.Scan(&Book.ID, &Book.ISBN, &Book.Title, &Book.Cover, &Book.Author, &Book.Published, &Book.Pages); err != nil {
			return models.Book{}, err
		}
	} else {
		return models.Book{}, fmt.Errorf("book wit id %d is not exist", book_id)
	}

	return Book, nil
}

func DeleteBook(user_id, book_id int) (books []models.BookStatus, err error) {

	// delete book from user books
	if _, err := Postgres.DB.Exec(`DELETE FROM user_books WHERE user_id=$1 and book_id=$2`, user_id, book_id); err != nil {
		return nil, err
	}

	// check, if book useless then delete
	rows, err := Postgres.DB.Query(fmt.Sprintf(`SELECT * FROM user_books where book_id=%d`, book_id))
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		if _, err := Postgres.DB.Exec(`DELETE FROM books WHERE id=$1`, book_id); err != nil {
			return nil, err
		}
	}

	// return all books for user
	rows, err = Postgres.DB.Query(fmt.Sprintf(`SELECT id, isbn, title, cover, author, published, pages, status FROM books INNER JOIN user_books ON books.id=user_books.book_id WHERE user_id=%d`, user_id))
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
