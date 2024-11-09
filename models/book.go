package models

// type Book struct {
// 	Author       string    `json:"author"`
// 	Pages        int       `json:"pages"`
// 	Published_at time.Time `json:"published_at"`
// 	Title        string    `json:"title"`
// }

type Book struct {
	Title  string `firestore:"title"`
	Author string `firestore:"author"`
	Pages  int    `firestore:"pages"`
}
