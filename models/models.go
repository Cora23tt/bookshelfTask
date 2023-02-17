package models

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type Book struct {
	ID        int    `json:"id"`
	ISBN      string `json:"isbn"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	Author    string `json:"author"`
	Published int    `json:"published"`
	Pages     int    `json:"pages"`
}

type BookResponse struct {
	Data    BookStatus `json:"data"`
	IsOK    bool       `json:"isOk"`
	Message string     `json:"message"`
}

type BooksResponse struct {
	Data    []BookStatus `json:"data"`
	IsOK    bool         `json:"isOk"`
	Message string       `json:"message"`
}

type BookStatus struct {
	Book   Book `json:"book"`
	Status int  `json:"status"`
}

type ISBN struct {
	ISBN string `json:"isbn"`
}

type Status struct {
	Status int `json:"status"`
}

type BookPublishDate struct {
	PublishDate string `json:"publish_date"`
}

type Authors struct {
	Authors []DescribeURL `json:"authors"`
}

type Author struct {
	Name string `json:"personal_name"`
}

type DescribeURL struct {
	URL string `json:"key"`
}

type BookTtitle struct {
	Title string `json:"title"`
}

type ISBN10 struct {
	ISBN10 string `json:"isbn_10"`
}

type ISBN13 struct {
	ISBN13 string `json:"isbn_13"`
}

type BookCovers struct {
	Covers []int `json:"covers"`
}

type NumberOfPages struct {
	Pages int `json:"number_of_pages"`
}

// JSON response from openlibrary.org
// {
// 	"type": {
// 		"key": "/type/edition"
// 	},
// 	"publish_date": "2008",							// Published
// 	"publish_country": "xx",
// 	"languages": [{"key": "/languages/eng"}],
// 	"oclc_numbers": ["264037956"],

// 	"authors": [{"key": "/authors/OL10430104A"}], 	// AUTHORS

// 	"title": "Alice in the Know", 					// TITLE

// 	"publishers": ["Paw Prints"],

// 	"isbn_13": ["9781439574164"], 					// ISBN 13
// 	"isbn_10": ["1439574162"], 						// ISBN 10
// 	"pagination": "pages",
// 	"ocaid": "aliceinknow0000unse",
// 	"source_records": ["ia:aliceinknow0000unse"],
// 	"full_title": "Alice in the Know",
// 	"covers": [12960596], 							// COVERS
// 	"works": [{"key": "/works/OL29230866W"}],
// 	"key": "/books/OL40206746M",
// 	"latest_revision": 1,
// 	"revision": 1,
// 	"created": {
// 		"type": "/type/datetime",
// 		"value": "2022-10-20T12:47:43.422828"
// 	},
// 	"last_modified": {
// 		"type": "/type/datetime",
// 		"value": "2022-10-20T12:47:43.422828"
// 	}
// }

// JSON response from openlibrary.org
// {
// 	"publishers": ["Alfaguara"],
// 	"number_of_pages": 160,
// 	"weight": "6.1 ounces",
// 	"covers": [2269492],
// 	"physical_format": "Paperback",
// 	"key": "/books/OL9546503M",
// 	"authors": [{"key": "/authors/OL34184A"}, {"key": "/authors/OL2833811A"}],
// 	"subjects": ["Juvenile Fiction - Classics", "Humorous Stories", "Children's 9-12 - Fiction - General", "Spanish: Young Adult (Gr. 7-9)"],
// 	"isbn_13": ["9788420431024"],
// 	"title": "El superzorro",
// 	"identifiers": {
// 		"librarything": ["6446"],
// 		"goodreads": ["1507550"]
// 	},
// 	"edition_name": "5a ed edition",
// 	"isbn_10": ["8420431028"],
// 	"publish_date": "December 1995",
// 	"works": [{"key": "/works/OL45804W"}],
// 	"type": {"key": "/type/edition"},
// 	"physical_dimensions": "8 x 4.7 x 0.4 inches",
// 	"latest_revision": 8,
// 	"revision": 8,
// 	"created": {
// 		"type": "/type/datetime",
// 		"value": "2008-04-30T09:38:13.731961"
// 	},
// 	"last_modified": {
// 		"type": "/type/datetime",
// 		"value": "2022-09-09T18:07:44.617112"
// 	}
// }
