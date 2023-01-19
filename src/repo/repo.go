package repo

import "github.com/gofiber/fiber/v2"

type Repo interface {
	Start(router fiber.Router) error
}
