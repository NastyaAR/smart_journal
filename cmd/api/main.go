package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"blockchain_project/internal/adapters"
	"blockchain_project/internal/adapters/mocks"
	"blockchain_project/internal/handlers"
	"blockchain_project/internal/repositories"
	"blockchain_project/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "Smart Journal API",
		ServerHeader: "Fiber",
	})

	app.Use(cors.New())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://admin:password@0.0.0.0:5432/smartjournal?sslmode=disable"
	}
	postgresRepo, err := repositories.NewPostgresRepository(ctx, connString)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer postgresRepo.Close()

	contractAdapter, err := adapters.NewContractAdapter(
		os.Getenv("RPC_URL"),
		os.Getenv("CONTRACT_ADDRESS"),
	)
	if err != nil {
		log.Printf("Failed to create contract adapter: %v", err)
		contractAdapter = &mocks.ContractAdapterMock{}
	}

	teacherRepo := repositories.NewTeacherRepository(postgresRepo)
	studentRepo := repositories.NewStudentRepository(postgresRepo)
	groupRepo := repositories.NewGroupRepository(postgresRepo)
	subjectRepo := repositories.NewSubjectRepository(postgresRepo)
	gradeRepo := repositories.NewGradeRepository(postgresRepo)
	achievementRepo := repositories.NewAchievementRepository(postgresRepo)
	merchRepo := repositories.NewMerchRepository(postgresRepo)
	recommendationRepo := repositories.NewRecommendationRepository(postgresRepo)
	userRepository := repositories.NewUserRepository(postgresRepo)

	studentService := services.NewStudentService(studentRepo, achievementRepo, userRepository, contractAdapter)
	teacherService := services.NewTeacherService(teacherRepo, achievementRepo, studentService, groupRepo, subjectRepo, gradeRepo, studentRepo, userRepository)

	aiService := services.NewAIService(os.Getenv("AI_SERVICE_URL"))

	authHandler := handlers.NewAuthHandler(services.NewAuthService(userRepository))
	studentHandler := handlers.NewStudentHandler(
		studentService,
		achievementRepo,
		gradeRepo,
		groupRepo,
		subjectRepo,
		merchRepo,
		recommendationRepo,
		aiService,
		authHandler,
	)
	teacherHandler := handlers.NewTeacherHandler(teacherService, authHandler)

	setupRoutes(app, teacherRepo, studentRepo, groupRepo, subjectRepo, gradeRepo, achievementRepo, merchRepo, studentService, authHandler, teacherHandler, studentHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server: ", err)
	}
}

func setupRoutes(
	app *fiber.App,
	teacherRepo *repositories.TeacherRepository,
	studentRepo *repositories.StudentRepository,
	groupRepo *repositories.GroupRepository,
	subjectRepo *repositories.SubjectRepository,
	gradeRepo *repositories.GradeRepository,
	achievementRepo *repositories.AchievementRepository,
	merchRepo *repositories.MerchRepository,
	studentService *services.StudentService,
	authHandler *handlers.AuthHandler,
	teacherHandler *handlers.TeacherHandler,
	studentHandler *handlers.StudentHandler,
) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	auth := app.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/session", authHandler.GetSession)

	teachers := app.Group("/teachers", authHandler.RequireAuth, authHandler.RequireRole("teacher"))
	teachers.Post("/achievements/confirm", teacherHandler.ConfirmAchievement)
	teachers.Post("/achievements/deny", teacherHandler.DenyAchievement)
	teachers.Post("/tokens/award", teacherHandler.AwardTokensManually)
	teachers.Get("/groups", teacherHandler.GetGroups)
	teachers.Post("/groups", teacherHandler.CreateGroup)
	teachers.Post("/groups/attach", teacherHandler.AttachGroup)
	teachers.Get("/groups/:id/grades", teacherHandler.GetGradesForGroup)
	teachers.Post("/subjects", teacherHandler.CreateSubject)
	teachers.Post("/subjects/attach", teacherHandler.AttachSubject)
	teachers.Post("/groups/add-student", teacherHandler.AddStudentToGroup)
	teachers.Post("/grades", teacherHandler.SetGrade)
	teachers.Get("/achievements/pending", teacherHandler.GetPendingAchievements)

	app.Post("/register", studentHandler.Register)

	students := app.Group("/students", studentHandler.RequireAuth)
	students.Post("/achievements", studentHandler.CreateAchievement)
	students.Get("/achievements", studentHandler.GetAchievements)
	students.Get("/grades", studentHandler.GetGrades)
	students.Get("/group", studentHandler.GetGroup)
	students.Get("/balance", studentHandler.GetBalance)
	students.Get("/merch", studentHandler.GetMerch)
	students.Post("/merch/buy", studentHandler.BuyMerch)
	students.Get("/purchases", studentHandler.GetPurchases)
	students.Post("/recommendations", studentHandler.GenerateRecommendations)
	students.Get("/recommendations", studentHandler.GetLatestRecommendation)
}
