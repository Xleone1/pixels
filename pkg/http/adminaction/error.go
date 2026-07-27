package adminaction

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// HTTPError maps expected administrative action failures.
func HTTPError(err error) error {
	if errors.Is(err, ErrInvalidAudit) {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidAudit.Error())
	}
	if errors.Is(err, ErrForbidden) {
		return fiber.NewError(fiber.StatusForbidden, ErrForbidden.Error())
	}
	return err
}
