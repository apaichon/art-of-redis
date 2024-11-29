// main.go
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Question struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Correct int      `json:"correct"`
}

type Quiz struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

var ctx = context.Background()
var rdb *redis.Client

func main() {
	// Initialize Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	// Initialize Fiber with template engine
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Serve static files
	app.Static("/", "./public")

	// Routes
	app.Get("/", handleHome)
	app.Get("/api/quiz/:id", handleGetQuiz)
	app.Post("/api/quiz", handleCreateQuiz)
	app.Post("/api/submit/:id", handleSubmitQuiz)

	log.Fatal(app.Listen(":3000"))
}

func handleHome(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{})
}

func handleGetQuiz(c *fiber.Ctx) error {
	id := c.Params("id")
	val, err := rdb.Get(ctx, "quiz:"+id).Result()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz not found"})
	}

	var quiz Quiz
	json.Unmarshal([]byte(val), &quiz)
	return c.JSON(quiz)
}

func handleCreateQuiz(c *fiber.Ctx) error {
	var quiz Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return err
	}

	// Store quiz in Redis
	quizJson, _ := json.Marshal(quiz)
	err := rdb.Set(ctx, "quiz:"+quiz.ID, quizJson, 0).Err()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save quiz"})
	}

	return c.JSON(quiz)
}

func handleSubmitQuiz(c *fiber.Ctx) error {
	type Submission struct {
		Answers []int `json:"answers"`
	}

	var submission Submission
	if err := c.BodyParser(&submission); err != nil {
		return err
	}

	// Get quiz from Redis
	quizId := c.Params("id")
	val, err := rdb.Get(ctx, "quiz:"+quizId).Result()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Quiz not found"})
	}

	var quiz Quiz
	json.Unmarshal([]byte(val), &quiz)

	// Calculate score
	score := 0
	for i, answer := range submission.Answers {
		if i < len(quiz.Questions) && answer == quiz.Questions[i].Correct {
			score++
		}
	}

	return c.JSON(fiber.Map{
		"score": score,
		"total": len(quiz.Questions),
	})
}
