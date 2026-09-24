package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ratuhin1122/URL-shortner/database"
	"github.com/ratuhin1122/URL-shortner/helpers"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"custom_short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit_remaining"`
	XRateLimitReset int64         `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(Request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	// implement rate limiting
	r1 := database.CreateClient(1)
	defer r1.Close()

	value, err := r1.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// first time this IP is seen — set quota with 30 min TTL
		_ = r1.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "can't connect to server",
		})
	} else {
		valueInt, _ := strconv.Atoi(value)
		if valueInt <= 0 {
			limit, _ := r1.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// check domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid domain",
		})
	}

	// check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL",
		})
	}

	// enforce HTTPS
	body.URL = helpers.EnforceHTTP(body.URL)

	// generate or use custom short code
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// check if short code already exists in DB 0
	r0 := database.CreateClient(0)
	defer r0.Close()

	val, _ := r0.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL custom short is already in use",
		})
	}

	// default expiry = 24 hours
	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}

	// save short code → original URL in Redis DB 0
	err = r0.Set(database.Ctx, id, body.URL, body.Expiry*time.Hour).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to connect to server",
		})
	}

	// build response
	resp := Response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// decrement the rate limit counter for this IP
	r1.Decr(database.Ctx, c.IP())

	// get updated remaining quota
	val, _ = r1.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	// get TTL for rate limit reset time
	ttl, _ := r1.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = int64(ttl / time.Nanosecond / time.Minute)

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusCreated).JSON(resp)
}