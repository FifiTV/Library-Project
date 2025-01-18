package models

type ProposedBook struct {
	Title    string    `firestore:"title"`
	Author   string    `firestore:"author"`
	Comments []Comment `firestore:"comments"`
	UpVoutes int       `firestore:"upVotes"`
}
