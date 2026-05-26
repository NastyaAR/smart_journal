package services

import (
	"context"
	"fmt"

	"blockchain_project/internal/models"
	"blockchain_project/internal/repositories"

	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TeacherService struct {
	repo               *repositories.TeacherRepository
	achievementRepo    *repositories.AchievementRepository
	studentService     repositories.StudentServiceManager
	groupRepo          *repositories.GroupRepository
	subjectRepo        *repositories.SubjectRepository
	gradeRepo          *repositories.GradeRepository
	studentRepo        *repositories.StudentRepository
	tokenOperationRepo *repositories.TokenOperationRepository
	userRepository     *repositories.UserRepository
}

func NewTeacherService(
	repo *repositories.TeacherRepository,
	achievementRepo *repositories.AchievementRepository,
	studentService repositories.StudentServiceManager,
	groupRepo *repositories.GroupRepository,
	subjectRepo *repositories.SubjectRepository,
	gradeRepo *repositories.GradeRepository,
	studentRepo *repositories.StudentRepository,
	tokenOperationRepo *repositories.TokenOperationRepository,
	userRepository *repositories.UserRepository,
) *TeacherService {
	return &TeacherService{
		repo:               repo,
		achievementRepo:    achievementRepo,
		studentService:     studentService,
		groupRepo:          groupRepo,
		subjectRepo:        subjectRepo,
		gradeRepo:          gradeRepo,
		studentRepo:        studentRepo,
		tokenOperationRepo: tokenOperationRepo,
		userRepository:     userRepository,
	}
}

func (s *TeacherService) Authenticate(ctx context.Context, email, password string) (*models.Teacher, error) {
	user, err := s.userRepository.Authenticate(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	teacher, err := s.repo.GetTeacherByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher")
	}

	return teacher, nil
}

func (s *TeacherService) ConfirmAchievement(ctx context.Context, achievementID, teacherID int) error {
	achievement, err := s.achievementRepo.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}
	student, err := s.studentRepo.GetStudentByID(ctx, achievement.StudentID)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if err := s.requireGroupAccess(ctx, teacherID, student.GroupID); err != nil {
		return err
	}

	if err := s.achievementRepo.ConfirmAchievement(ctx, achievementID, teacherID); err != nil {
		return fmt.Errorf("failed to confirm achievement: %w", err)
	}

	if err := s.studentService.AwardTokensForAchievement(ctx, achievementID); err != nil {
		return fmt.Errorf("failed to award tokens: %w", err)
	}

	return nil
}

func (s *TeacherService) AwardTokensManually(ctx context.Context, studentID, teacherID, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if err := s.requireGroupAccess(ctx, teacherID, student.GroupID); err != nil {
		return err
	}

	if _, err := s.studentService.AddStudentTokens(ctx, studentID, amount); err != nil {
		return fmt.Errorf("failed to update student tokens: %w", err)
	}

	points := new(big.Int).SetInt64(int64(amount))
	studentAddress := common.HexToAddress(fmt.Sprintf("0x%040d", studentID))
	if err := s.studentService.GetContractAdapter().AwardTokens(studentAddress, points); err != nil {
		return fmt.Errorf("failed to award tokens on blockchain: %w", err)
	}

	if s.tokenOperationRepo != nil {
		if err := s.tokenOperationRepo.Create(ctx, &models.TokenOperation{
			StudentID:     studentID,
			TeacherID:     &teacherID,
			Amount:        amount,
			OperationType: "manual_award",
			Reason:        "Manual award",
		}); err != nil {
			return fmt.Errorf("failed to record token operation: %w", err)
		}
	}

	return nil
}

func (s *TeacherService) CreateGroup(ctx context.Context, teacherID int, group *models.Group) error {
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		return err
	}
	return s.repo.AssignGroup(ctx, teacherID, group.ID)
}

func (s *TeacherService) CreateSubject(ctx context.Context, teacherID int, subject *models.Subject) error {
	if err := s.subjectRepo.CreateSubject(ctx, subject); err != nil {
		return err
	}
	return s.repo.AssignSubject(ctx, teacherID, subject.ID)
}

func (s *TeacherService) AddStudentToGroup(ctx context.Context, teacherID, studentID, groupID int) error {
	if err := s.requireGroupAccess(ctx, teacherID, groupID); err != nil {
		return err
	}

	_, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}

	_, err = s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	if err := s.studentRepo.AddToGroup(ctx, studentID, groupID); err != nil {
		return fmt.Errorf("failed to add student to group: %w", err)
	}

	return nil
}

func (s *TeacherService) SetGrade(ctx context.Context, teacherID int, grade *models.Grade) error {
	if grade.Value < 1 || grade.Value > 5 {
		return fmt.Errorf("grade value must be between 1 and 5")
	}
	student, err := s.studentRepo.GetStudentByID(ctx, grade.StudentID)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if err := s.requireGroupAccess(ctx, teacherID, student.GroupID); err != nil {
		return err
	}
	if _, err := s.subjectRepo.GetSubjectByID(ctx, grade.SubjectID); err != nil {
		return fmt.Errorf("failed to get subject: %w", err)
	}
	if err := s.requireSubjectAccess(ctx, teacherID, grade.SubjectID); err != nil {
		return err
	}
	return s.gradeRepo.CreateGrade(ctx, grade)
}

func (s *TeacherService) DenyAchievement(ctx context.Context, achievementID, teacherID int) error {
	achievement, err := s.achievementRepo.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}
	student, err := s.studentRepo.GetStudentByID(ctx, achievement.StudentID)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if err := s.requireGroupAccess(ctx, teacherID, student.GroupID); err != nil {
		return err
	}

	return s.achievementRepo.DenyAchievement(ctx, achievementID, teacherID)
}

func (s *TeacherService) GetPendingAchievements(ctx context.Context, teacherID int) ([]*models.PendingAchievementView, error) {
	return s.achievementRepo.GetPendingAchievements(ctx, teacherID)
}

func (s *TeacherService) GetTeacherByUserID(ctx context.Context, userID int) (*models.Teacher, error) {
	return s.repo.GetTeacherByUserID(ctx, userID)
}

func (s *TeacherService) GetGroups(ctx context.Context, teacherID int) ([]*models.Group, error) {
	return s.groupRepo.GetGroupsByTeacherID(ctx, teacherID)
}

func (s *TeacherService) AttachGroup(ctx context.Context, teacherID, groupID int) error {
	if _, err := s.groupRepo.GetGroupByID(ctx, groupID); err != nil {
		return err
	}
	return s.repo.AssignGroup(ctx, teacherID, groupID)
}

func (s *TeacherService) AttachSubject(ctx context.Context, teacherID, subjectID int) error {
	if _, err := s.subjectRepo.GetSubjectByID(ctx, subjectID); err != nil {
		return err
	}
	return s.repo.AssignSubject(ctx, teacherID, subjectID)
}

func (s *TeacherService) GetSubjects(ctx context.Context, teacherID int) ([]*models.Subject, error) {
	return s.subjectRepo.GetSubjectsByTeacherID(ctx, teacherID)
}

func (s *TeacherService) GetStudentsByGroupID(ctx context.Context, teacherID, groupID int) ([]*models.Student, error) {
	if err := s.requireGroupAccess(ctx, teacherID, groupID); err != nil {
		return nil, err
	}
	return s.studentRepo.GetStudentsByGroupID(ctx, groupID)
}

func (s *TeacherService) GetGradeViewsByGroupID(ctx context.Context, teacherID, groupID int) ([]*models.GradeView, error) {
	if _, err := s.groupRepo.GetGroupByID(ctx, groupID); err != nil {
		return nil, err
	}
	if err := s.requireGroupAccess(ctx, teacherID, groupID); err != nil {
		return nil, err
	}
	return s.gradeRepo.GetGradeViewsByGroupID(ctx, groupID)
}

func (s *TeacherService) GetTokenOperationsByGroupID(ctx context.Context, teacherID, groupID int) ([]*models.TokenOperation, error) {
	if err := s.requireGroupAccess(ctx, teacherID, groupID); err != nil {
		return nil, err
	}
	if s.tokenOperationRepo == nil {
		return []*models.TokenOperation{}, nil
	}
	return s.tokenOperationRepo.GetByGroupID(ctx, groupID)
}

func (s *TeacherService) requireGroupAccess(ctx context.Context, teacherID, groupID int) error {
	if groupID == 0 {
		return fmt.Errorf("student is not assigned to a group")
	}
	ok, err := s.repo.HasGroup(ctx, teacherID, groupID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("teacher has no access to this group")
	}
	return nil
}

func (s *TeacherService) requireSubjectAccess(ctx context.Context, teacherID, subjectID int) error {
	ok, err := s.repo.HasSubject(ctx, teacherID, subjectID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("teacher has no access to this subject")
	}
	return nil
}
