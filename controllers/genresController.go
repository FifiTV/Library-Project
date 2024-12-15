package controllers

import (
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/iterator"
)

func GetGenres(c *fiber.Ctx) *[]models.Genre {
	var genres []models.Genre
	query := initializers.Client.Collection("booksGenres")

	iter := query.Documents(c.Context())
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil
		}
		var genre models.Genre
		if err := doc.DataTo(&genre); err != nil {
			return nil
		}
		genres = append(genres, genre)
	}

	return &genres
}
