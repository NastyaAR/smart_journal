package handlers

import (
	"context"
	"net/http"
	"time"

	"blockchain_project/internal/models"
	"blockchain_project/internal/services"

	"github.com/gofiber/fiber/v2"
)

type TeacherHandler struct {
	teacherService *services.TeacherService
	authHandler    *AuthHandler
}

func NewTeacherHandler(teacherService *services.TeacherService, authHandler *AuthHandler) *TeacherHandler {
	return &TeacherHandler{
		teacherService: teacherService,
		authHandler:    authHandler,
	}
}

func (h *TeacherHandler) currentTeacher(c *fiber.Ctx, ctx context.Context) (*models.Teacher, error) {
	sess, err := authStore.Get(c)
	if err != nil {
		return nil, err
	}

	userID, ok := sess.Get("user_id").(int)
	if !ok {
		return nil, fiber.NewError(http.StatusUnauthorized, "Teacher not authenticated")
	}

	role, ok := sess.Get("role").(string)
	if !ok || role != "teacher" {
		return nil, fiber.NewError(http.StatusForbidden, "Insufficient permissions")
	}

	return h.teacherService.GetTeacherByUserID(ctx, userID)
}

func (h *TeacherHandler) ConfirmAchievement(c *fiber.Ctx) error {
	var req struct {
		AchievementID int `json:"achievement_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.ConfirmAchievement(ctx, req.AchievementID, teacher.ID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Achievement confirmed successfully"})
}

func (h *TeacherHandler) DenyAchievement(c *fiber.Ctx) error {
	var req struct {
		AchievementID int `json:"achievement_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.DenyAchievement(ctx, req.AchievementID, teacher.ID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Achievement denied successfully"})
}

func (h *TeacherHandler) AwardTokensManually(c *fiber.Ctx) error {
	var req struct {
		StudentID int `json:"student_id"`
		Amount    int `json:"amount"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.AwardTokensManually(ctx, req.StudentID, teacher.ID, req.Amount); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Tokens awarded successfully"})
}

func (h *TeacherHandler) CreateGroup(c *fiber.Ctx) error {
	var group models.Group
	if err := c.BodyParser(&group); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.CreateGroup(ctx, teacher.ID, &group); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(group)
}

func (h *TeacherHandler) GetGroups(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	groups, err := h.teacherService.GetGroups(ctx, teacher.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(groups)
}

func (h *TeacherHandler) AttachGroup(c *fiber.Ctx) error {
	var req struct {
		GroupID int `json:"group_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.AttachGroup(ctx, teacher.ID, req.GroupID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Group attached successfully"})
}

func (h *TeacherHandler) CreateSubject(c *fiber.Ctx) error {
	var subject models.Subject
	if err := c.BodyParser(&subject); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.CreateSubject(ctx, teacher.ID, &subject); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(subject)
}

func (h *TeacherHandler) AttachSubject(c *fiber.Ctx) error {
	var req struct {
		SubjectID int `json:"subject_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.AttachSubject(ctx, teacher.ID, req.SubjectID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Subject attached successfully"})
}

func (h *TeacherHandler) GetSubjects(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	subjects, err := h.teacherService.GetSubjects(ctx, teacher.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(subjects)
}

func (h *TeacherHandler) AddStudentToGroup(c *fiber.Ctx) error {
	var req struct {
		StudentID int `json:"student_id"`
		GroupID   int `json:"group_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.AddStudentToGroup(ctx, teacher.ID, req.StudentID, req.GroupID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Student added to group successfully"})
}

func (h *TeacherHandler) SetGrade(c *fiber.Ctx) error {
	var grade models.Grade
	if err := c.BodyParser(&grade); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if err := h.teacherService.SetGrade(ctx, teacher.ID, &grade); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(grade)
}

func (h *TeacherHandler) GetStudentsForGroup(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	students, err := h.teacherService.GetStudentsByGroupID(ctx, teacher.ID, groupID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(students)
}

func (h *TeacherHandler) GetGradesForGroup(c *fiber.Ctx) error {
	groupID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	grades, err := h.teacherService.GetGradeViewsByGroupID(ctx, teacher.ID, groupID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(grades)
}

func (h *TeacherHandler) GetPendingAchievements(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	teacher, err := h.currentTeacher(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	achievements, err := h.teacherService.GetPendingAchievements(ctx, teacher.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(achievements)
}
