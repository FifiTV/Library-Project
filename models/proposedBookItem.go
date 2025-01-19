package models

type PropsedBookItem struct {
	Title    string
	Author   string
	UpVotes  int
	Comments []CommentItem
}
