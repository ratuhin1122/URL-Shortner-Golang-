package routes

import (
	"os"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/ratuhin1122/URL-shortner/database"
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
	r2 := database.CreateClient(1)
	defer r2.Close()

	value, error := r2.Get(database.Ctx, c.IP()).Result()
	if error == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"),30*60*time.Second).Err()
	}else{
		value, _ = r2.Get(database.Ctx, c.IP()).Result()
		valueInt = strconv.Atoi(value)
		if valueInt <= 0{
			limit, _ = r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "rate limit exceeded",
				"retry_after": limit / time.Nanosecond / time.Minute,
			})
		} 
	}

	


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
	var id string
	if body.CustomShort == ""{
		id = uuid.New().String()[:6]
	}else{
		id = body.CustomShort
	}
	r2 := database.CreateClient(0)
	defer r2.Close()
	val _ = r2.Get(database.Ctx, c.IP()).Result()
	if val != ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL already exists"
		})
	}
	if body.Expiry == 0{
		body.Expiry = 24
	}

	err = r2.Set(database.Ctx, id, body.URL, body.Expiry).Err()
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid URL"
		})
	}

	resp := Response{
		URL: body.URL,
		CustomShort: "",
		Expiry: body.Expiry,
		XRateRemaining: 10,
		XRateLimitReset: 30,
	}
	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx,c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ = r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusCreated).JSON(resp)






	// TODO: store in Redis and return response
	return nil
}