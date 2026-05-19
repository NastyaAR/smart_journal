package handlers

import (
	"context"
	"net/http"
	"time"

	"blockchain_project/internal/models"
	"blockchain_project/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var authStore = session.New(session.Config{
	Expiration: 24 * time.Hour,
})

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func handleFiberError(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
	}
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.authService.Authenticate(ctx, req.Login, req.Password)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	sess, err := authStore.Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
	}

	sess.Set("user_id", user.ID)
	sess.Set("role", user.Role)
	if err := sess.Save(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save session"})
	}

	return c.JSON(models.LoginResponse{
		Message: "Login successful",
		Role:    user.Role,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, err := authStore.Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	if err := sess.Destroy(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to destroy session"})
	}

	return c.JSON(fiber.Map{"message": "Logout successful"})
}

func (h *AuthHandler) GetSession(c *fiber.Ctx) error {
	sess, err := authStore.Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
	}

	return c.JSON(fiber.Map{
		"user_id": userID,
		"role":    sess.Get("role"),
	})
}

func (h *AuthHandler) RequireAuth(c *fiber.Ctx) error {
	sess, err := authStore.Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	if sess.Get("user_id") == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
	}

	return c.Next()
}

func (h *AuthHandler) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := authStore.Get(c)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
		}

		role := sess.Get("role")
		if role == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}

		roleStr, ok := role.(string)
		if !ok {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid role type"})
		}

		for _, r := range roles {
			if r == roleStr {
				return c.Next()
			}
		}

		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}
}
