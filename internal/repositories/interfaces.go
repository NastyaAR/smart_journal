package repositories

import (
	"blockchain_project/internal/adapters"
	"context"
)

type StudentServiceManager interface {
	AwardTokensForAchievement(ctx context.Context, achievementID int) error
	AddStudentTokens(ctx context.Context, id int, amount int) (int, error)
	GetContractAdapter() adapters.ContractAdapter
}
