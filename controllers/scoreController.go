package controllers

import (
	"context"
	"fmt"
	"my-firebase-project/initializers"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetBookScore(c *fiber.Ctx) error {
	userId := c.Params("user")
	bookId := c.Params("book")

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	bookIdInt, err := strconv.Atoi(bookId)
	if err != nil {
		return err
	}

	client := initializers.Client
	ref := client.Collection("booksRatings").Doc(fmt.Sprintf("%d-%d", userIdInt, bookIdInt))

	doc, err := ref.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.JSON(fiber.Map{
				"status":  "error",
				"message": "Rating not found",
			})
		}
		return err
	}

	if !doc.Exists() {
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Rating not found",
		})
	}

	var result struct {
		Rating int `json:"rating"`
	}

	doc.DataTo(&result)

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func ScoreBook(c *fiber.Ctx) error {
	userIdStr := c.Params("user")
	bookIdStr := c.Params("book")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return err
	}
	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		return err
	}

	var body struct {
		Rating int `json:"rating"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	client := initializers.Client
	ref := client.Collection("booksRatings").Doc(fmt.Sprintf("%d-%d", userId, bookId))

	doc, err := ref.Get(context.Background())
	if err != nil && status.Code(err).String() != codes.NotFound.String() {
		return err
	}

	if doc.Exists() {
		_, err = ref.Update(context.Background(), []firestore.Update{
			{
				Path:  "rating",
				Value: body.Rating,
			},
		})
	} else {
		_, err = ref.Set(context.Background(), map[string]interface{}{
			"userId": userId,
			"bookId": bookId,
			"rating": body.Rating,
		})
	}

	if err != nil {
		return err
	}

	// sc, ll, _ := GetAvgScore(bookId)
	UpdateBookAvgScore(bookId)
	// Respond with success
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Rating updated/added successfully",
	})
}

func GetAvgScore(bookId int) (float64, int, error) {
	client := initializers.Client

	ratingsQuery := client.Collection("booksRatings").Where("bookId", "==", bookId)
	ratingsDocs, err := ratingsQuery.Documents(context.Background()).GetAll()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch ratings: %v", err)
	}

	if len(ratingsDocs) == 0 {
		return 0, 0, nil
	}

	var totalRatings int
	for _, rDoc := range ratingsDocs {
		var ratingData struct {
			Rating int `firestore:"rating"`
		}
		if err := rDoc.DataTo(&ratingData); err != nil {
			return 0, 0, fmt.Errorf("failed to parse rating data: %v", err)
		}
		totalRatings += ratingData.Rating
	}

	avgScore := float64(totalRatings) / float64(len(ratingsDocs))

	return avgScore, len(ratingsDocs), nil
}

func UpdateBookAvgScore(bookId int) error {
	client := initializers.Client

	avgScore, count, err := GetAvgScore(bookId)
	if err != nil {
		return fmt.Errorf("failed to calculate avgScore: %v", err)
	}

	if count == 0 {
		avgScore = 0
	}

	booksQuery := client.Collection("books").Where("id", "==", bookId).Limit(1)
	booksDocs, err := booksQuery.Documents(context.Background()).GetAll()
	if err != nil {
		return fmt.Errorf("failed to fetch book with id %d: %v", bookId, err)
	}

	if len(booksDocs) == 0 {
		return fmt.Errorf("book with id %d not found", bookId)
	}

	bookDocRef := booksDocs[0].Ref
	_, err = bookDocRef.Update(context.Background(), []firestore.Update{
		{
			Path:  "avg_score",
			Value: avgScore,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update avgScore for book id %d: %v", bookId, err)
	}

	return nil
}
