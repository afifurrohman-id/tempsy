package auth

import (
	"github.com/gofiber/fiber/v2"
)

var AllowedHttpMethod = []string{fiber.MethodGet, fiber.MethodDelete, fiber.MethodOptions, fiber.MethodPut, fiber.MethodPost}

const (
	BearerPrefix = "Bearer "
	// follow HTTP 2.0 (lowercase), but still backward compatible with HTTP 1.1 because it's case insensitive
	HeaderRealIp  = "real-IP"
	HeaderXRealIp = "x-real-ip" // Backward compatibility purpose
)
