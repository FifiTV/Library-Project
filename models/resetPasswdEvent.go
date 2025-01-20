package models

import (
	"time"
)

type ResetPasswdEvent struct {
	UserId   int       `firestore:"user_id"`
	Start    time.Time `firestore:"start"`
	End      time.Time `firestore:"end"`
	KeyValue string    `firestore:"key_value"`
}
