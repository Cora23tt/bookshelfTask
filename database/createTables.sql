--- psql -U postgres -d bookshelf -a -f createTables.sql
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(50) NOT NULL,
	email VARCHAR(50) NOT NULL,
	key VARCHAR(50) NOT NULL,
	secret VARCHAR(50) NOT NULL,
	UNIQUE(email),
	UNIQUE(name),
	UNIQUE(key)
);
	
CREATE TABLE IF NOT EXISTS books (
	id SERIAL PRIMARY KEY,
	isbn VARCHAR(13) NOT NULL,
	title VARCHAR(150) NOT NULL,
	cover VARCHAR(250) NOT NULL,
    	author VARCHAR(250) NOT NULL,
	published INT NOT NULL,
	pages INT NOT NULL,
	UNIQUE(isbn)
);

CREATE TABLE IF NOT EXISTS user_books (
	user_id INT NOT NULL,
	book_id INT NOT NULL,
	status INT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY (book_id) REFERENCES books(id));

