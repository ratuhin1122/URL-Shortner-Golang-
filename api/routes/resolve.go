package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/ratuhin1122/URL-shortner/database"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	rdb := database.CreateClient(0)
	defer rdb.Close()

	value, err := rdb.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "short URL not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "can't connect to server",
		})
	}

	// increment visit counter in DB 1
	rInr := database.CreateClient(1)
	defer rInr.Close()

	rInr.Incr(database.Ctx, "counter")

	return c.Redirect(value, 301)
}
