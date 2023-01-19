package main

import (
	"Galaxy/src/repo/maven"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joomcode/errorx"
)

func main() {

	srv := fiber.New(fiber.Config{StreamRequestBody: true})

	srv.Use(logger.New(logger.Config{
		Format: fmt.Sprintf("[${ip}]:${port} ${status} - ${method} ${path}\n"),
	}))

	mavenRepo := &maven.Maven{
		MainDir:  "data/maven/",
		RootDirs: []string{"rootDir"},
	}

	err := mavenRepo.Start(srv.Group("/maven"))
	if err != nil {
		panic(errorx.Panic(err))
	}

	/*
		srv.All("/maven", func(c *fiber.Ctx) error {

			// Print all request data along with url and protocol
			fmt.Println(string(c.Body()))
			fmt.Println()

			return c.SendString("Hello, World!")
		})

		srv.All("/docker", func(c *fiber.Ctx) error {

			// Print all request data along with url and protocol
			fmt.Println(string(c.Body()))
			fmt.Println()

			return c.SendString("Hello, World 1!")
		})

		srv.
			srv.Put("/releases/*", func(c *fiber.Ctx) error {

			// Print all request data along with url and protocol
			//fmt.Println(string(c.Body()))
			//fmt.Println()

			return c.SendString("Hello, World!")
		})*/

	err = srv.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
