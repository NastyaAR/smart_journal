package services

import (
	"context"
	"fmt"
	"math/big"

	"blockchain_project/internal/adapters"
	"blockchain_project/internal/models"
	"blockchain_project/internal/repositories"

	"github.com/ethereum/go-ethereum/common"
)

type StudentService struct {
	repo               *repositories.StudentRepository
	achievementRepo    *repositories.AchievementRepository
	tokenOperationRepo *repositories.TokenOperationRepository
	userRepository     *repositories.UserRepository
	contractAdapter    adapters.ContractAdapter
}

func NewStudentService(
	repo *repositories.StudentRepository,
	achievementRepo *repositories.AchievementRepository,
	tokenOperationRepo *repositories.TokenOperationRepository,
	userRepository *repositories.UserRepository,
	contractAdapter adapters.ContractAdapter,
) *StudentService {
	return &StudentService{
		repo:               repo,
		achievementRepo:    achievementRepo,
		tokenOperationRepo: tokenOperationRepo,
		userRepository:     userRepository,
		contractAdapter:    contractAdapter,
	}
}

func (s *StudentService) Authenticate(ctx context.Context, email, password string) (*models.Student, error) {
	user, err := s.userRepository.Authenticate(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	student, err := s.repo.GetStudentByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student")
	}

	return student, nil
}

func (s *StudentService) AwardTokensForAchievement(ctx context.Context, achievementID int) error {
	achievement, err := s.achievementRepo.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}
	if !achievement.Confirmed {
		return fmt.Errorf("achievement is not confirmed by teacher")
	}

	studentID := achievement.StudentID

	points := big.NewInt(10)
	if err := s.contractAdapter.AwardTokens(studentBlockchainAddress(studentID), points); err != nil {
		return fmt.Errorf("failed to award tokens: %w", err)
	}

	if _, err := s.repo.AddStudentTokens(ctx, studentID, 10); err != nil {
		return fmt.Errorf("failed to update student tokens in db: %w", err)
	}

	if s.tokenOperationRepo != nil {
		var teacherID *int
		if achievement.ConfirmedByTeacherID != 0 {
			value := achievement.ConfirmedByTeacherID
			teacherID = &value
		}
		if err := s.tokenOperationRepo.Create(ctx, &models.TokenOperation{
			StudentID:     studentID,
			TeacherID:     teacherID,
			Amount:        10,
			OperationType: "achievement_reward",
			Reason:        achievement.Title,
		}); err != nil {
			return fmt.Errorf("failed to record token operation: %w", err)
		}
	}

	return nil
}

func (s *StudentService) GetStudentBalance(ctx context.Context, studentID int) (*big.Int, error) {
	student, err := s.repo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return big.NewInt(int64(student.Tokens)), nil
}

func (s *StudentService) GetStudentByID(ctx context.Context, id int) (*models.Student, error) {
	return s.repo.GetStudentByID(ctx, id)
}

func (s *StudentService) UpdateStudentTokens(ctx context.Context, id int, tokens int) error {
	return s.repo.UpdateStudentTokens(ctx, id, tokens)
}

func (s *StudentService) AddStudentTokens(ctx context.Context, id int, amount int) (int, error) {
	return s.repo.AddStudentTokens(ctx, id, amount)
}

func (s *StudentService) SpendStudentTokens(ctx context.Context, id int, amount int) (int, error) {
	return s.repo.SpendStudentTokens(ctx, id, amount)
}

func (s *StudentService) PurchaseMerch(ctx context.Context, studentID, merchID, price int) (int, int, error) {
	points := big.NewInt(int64(price))
	if err := s.contractAdapter.RedeemTokens(studentBlockchainAddress(studentID), points); err != nil {
		return 0, 0, fmt.Errorf("failed to redeem tokens on blockchain: %w", err)
	}

	newBalance, purchaseID, err := s.repo.PurchaseMerch(ctx, studentID, merchID, price)
	if err != nil {
		return 0, 0, err
	}

	if s.tokenOperationRepo != nil {
		if err := s.tokenOperationRepo.Create(ctx, &models.TokenOperation{
			StudentID:     studentID,
			Amount:        -price,
			OperationType: "purchase",
			Reason:        fmt.Sprintf("Purchase #%d", purchaseID),
		}); err != nil {
			return 0, 0, fmt.Errorf("failed to record token operation: %w", err)
		}
	}

	return newBalance, purchaseID, nil
}

func (s *StudentService) GetStudentByUserID(ctx context.Context, userID int) (*models.Student, error) {
	return s.repo.GetStudentByUserID(ctx, userID)
}

func (s *StudentService) GetStudentsByGroupID(ctx context.Context, groupID int) ([]*models.Student, error) {
	return s.repo.GetStudentsByGroupID(ctx, groupID)
}

func (s *StudentService) GetTokenOperations(ctx context.Context, studentID int) ([]*models.TokenOperation, error) {
	if s.tokenOperationRepo == nil {
		return []*models.TokenOperation{}, nil
	}
	return s.tokenOperationRepo.GetByStudentID(ctx, studentID)
}

func (s *StudentService) CreateStudent(ctx context.Context, student *models.Student, userID int) error {
	return s.repo.CreateStudent(ctx, student, userID)
}

func (s *StudentService) CreateUser(ctx context.Context, user *models.User) error {
	return s.userRepository.CreateUser(ctx, user)
}

func (s *StudentService) DeleteStudent(ctx context.Context, id int) error {
	return s.repo.DeleteStudent(ctx, id)
}

func (s *StudentService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepository.DeleteUser(ctx, id)
}

func (s *StudentService) GetContractAdapter() adapters.ContractAdapter {
	return s.contractAdapter
}

func studentBlockchainAddress(studentID int) common.Address {
	return common.HexToAddress(fmt.Sprintf("0x%040d", studentID))
}
