package routes

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/ratuhin1122/URL-shortner/helpers"
)
type Request struct{
	URL           string        `json:"url"`
	CustomShort   string        `json:"custom_short"`
	Expiry        time.Duration `json:"expiry"`
}

type Response struct{
	URL                 string        `json:"url"`
	CustomShort         string        `json:"custom_short"`
	Expiry              time.Duration `json:"expiry"`
	XRateRemaining      int           `json:"rate_limit_remaining"`
	XRateLimitReset     int64         `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(Request)

	if err := c.BodyParser(&body); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request"
		})
	}

	//implement rate limiting

	// check domain error
	if !helpers.RemoveDomainError(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid domain"
		})
	}

	//chheck if the input is actual URL
	if !govalidator.IsURL(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL"
		})
	}

	//enforce HTTp, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	// TODO: store in Redis and return response
	return nil
}