package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func ListWorkflows(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"workflows": []interface{}{}})
}

func CreateWorkflow(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create workflow - TODO"})
}

func GetWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"id": id, "message": "Get workflow - TODO"})
}

func UpdateWorkflow(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update workflow - TODO"})
}

func DeleteWorkflow(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Delete workflow - TODO"})
}

func TriggerWorkflow(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Trigger workflow - TODO"})
}

func GetRun(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get run - TODO"})
}

func StreamRun(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Stream run - TODO"})
}
