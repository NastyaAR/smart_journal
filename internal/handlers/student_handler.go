package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blockchain_project/internal/models"
	"blockchain_project/internal/repositories"
	"blockchain_project/internal/services"

	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	studentService     *services.StudentService
	achievementRepo    *repositories.AchievementRepository
	gradeRepo          *repositories.GradeRepository
	groupRepo          *repositories.GroupRepository
	subjectRepo        *repositories.SubjectRepository
	merchRepo          *repositories.MerchRepository
	recommendationRepo *repositories.RecommendationRepository
	aiService          *services.AIService
	authHandler        *AuthHandler
}

func NewStudentHandler(
	studentService *services.StudentService,
	achievementRepo *repositories.AchievementRepository,
	gradeRepo *repositories.GradeRepository,
	groupRepo *repositories.GroupRepository,
	subjectRepo *repositories.SubjectRepository,
	merchRepo *repositories.MerchRepository,
	recommendationRepo *repositories.RecommendationRepository,
	aiService *services.AIService,
	authHandler *AuthHandler,
) *StudentHandler {
	return &StudentHandler{
		studentService:     studentService,
		achievementRepo:    achievementRepo,
		gradeRepo:          gradeRepo,
		groupRepo:          groupRepo,
		subjectRepo:        subjectRepo,
		merchRepo:          merchRepo,
		recommendationRepo: recommendationRepo,
		aiService:          aiService,
		authHandler:        authHandler,
	}
}

func (h *StudentHandler) currentStudent(c *fiber.Ctx, ctx context.Context) (*models.Student, error) {
	sess, err := authStore.Get(c)
	if err != nil {
		return nil, err
	}

	userID, ok := sess.Get("user_id").(int)
	if !ok {
		return nil, fiber.NewError(http.StatusUnauthorized, "Not authenticated")
	}

	role, ok := sess.Get("role").(string)
	if !ok || role != "student" {
		return nil, fiber.NewError(http.StatusForbidden, "Insufficient permissions")
	}

	return h.studentService.GetStudentByUserID(ctx, userID)
}

func (h *StudentHandler) Register(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		GroupID  int    `json:"group_id"`
		Password string `json:"password"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{
		Login:    req.Email,
		Password: req.Password,
		Role:     "student",
	}

	if err := h.studentService.CreateUser(ctx, &user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user account"})
	}

	student := models.Student{
		Name:    req.Name,
		Email:   req.Email,
		GroupID: req.GroupID,
		UserID:  user.ID,
	}

	if err := h.studentService.CreateStudent(ctx, &student, user.ID); err != nil {
		h.studentService.DeleteUser(ctx, user.ID)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Student registered successfully", "id": student.ID})
}

func (h *StudentHandler) GetGrades(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if student.GroupID == 0 {
		return c.JSON([]*models.GradeView{})
	}

	grades, err := h.gradeRepo.GetGradeViewsByGroupID(ctx, student.GroupID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(grades)
}

func (h *StudentHandler) GetGroup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	if student.GroupID == 0 {
		return c.JSON(fiber.Map{"group": nil, "students": []*models.Student{}})
	}

	group, err := h.groupRepo.GetGroupByID(ctx, student.GroupID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	studentsInGroup, err := h.studentService.GetStudentsByGroupID(ctx, student.GroupID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"group":    group,
		"students": studentsInGroup,
	})
}

func (h *StudentHandler) GetAchievements(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	achievements, err := h.achievementRepo.GetAchievementsByStudentID(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(achievements)
}

func (h *StudentHandler) CreateAchievement(c *fiber.Ctx) error {
	var achievement models.Achievement
	if err := c.BodyParser(&achievement); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	achievement.ID = 0
	achievement.StudentID = student.ID
	achievement.Status = "pending"
	achievement.Confirmed = false
	achievement.ConfirmedByTeacherID = 0

	if err := h.achievementRepo.CreateAchievement(ctx, &achievement); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(achievement)
}

func (h *StudentHandler) GetBalance(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	balance, err := h.studentService.GetStudentBalance(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"balance": balance.String()})
}

func (h *StudentHandler) GetMerch(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	merchList, err := h.merchRepo.GetAllMerch(ctx)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(merchList)
}

func (h *StudentHandler) BuyMerch(c *fiber.Ctx) error {
	var req struct {
		MerchID int `json:"merch_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	merch, err := h.merchRepo.GetMerchByID(ctx, req.MerchID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Merch not found"})
	}

	balance, err := h.studentService.GetStudentBalance(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if balance.Int64() < int64(merch.Price) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient tokens"})
	}

	newBalance, purchaseID, err := h.studentService.PurchaseMerch(ctx, student.ID, merch.ID, merch.Price)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Purchase successful",
		"merch":       merch,
		"purchase_id": purchaseID,
		"new_balance": newBalance,
	})
}

func (h *StudentHandler) GetPurchases(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	purchases, err := h.merchRepo.GetPurchasesByStudentID(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(purchases)
}

func (h *StudentHandler) GenerateRecommendations(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	grades, err := h.gradeRepo.GetGradeViewsByStudentID(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if len(grades) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "No grades found for recommendations"})
	}

	studentName, studentSurname := splitStudentName(student.Name)
	req := &models.AIRecommendationRequest{
		StudentID:      strconv.Itoa(student.ID),
		StudentName:    studentName,
		StudentSurname: studentSurname,
		Grades:         make([]models.AIGrade, 0, len(grades)),
	}
	for _, grade := range grades {
		req.Grades = append(req.Grades, models.AIGrade{
			Subject: grade.SubjectName,
			Score:   gradeScore(grade.Value),
		})
	}

	recommendation, err := h.aiService.GetRecommendations(ctx, req)
	if err != nil {
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	stored, err := h.recommendationRepo.CreateRecommendation(ctx, student.ID, recommendation)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(stored)
}

func (h *StudentHandler) GetLatestRecommendation(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	student, err := h.currentStudent(c, ctx)
	if err != nil {
		return handleFiberError(c, err)
	}

	recommendation, err := h.recommendationRepo.GetLatestByStudentID(ctx, student.ID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(recommendation)
}

func splitStudentName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func gradeScore(value int) int {
	if value <= 5 {
		return value * 20
	}
	return value
}

func (h *StudentHandler) RequireAuth(c *fiber.Ctx) error {
	sess, err := authStore.Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	if sess.Get("user_id") == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
	}

	role, ok := sess.Get("role").(string)
	if !ok || role != "student" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	return c.Next()
}
