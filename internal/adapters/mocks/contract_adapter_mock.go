package mocks

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type ContractAdapterMock struct {
	mock.Mock
}

func (m *ContractAdapterMock) AwardTokens(studentAddress common.Address, amount *big.Int) error {
	if len(m.ExpectedCalls) == 0 {
		return nil
	}
	args := m.Called(studentAddress, amount)
	return args.Error(0)
}

func (m *ContractAdapterMock) RedeemTokens(studentAddress common.Address, amount *big.Int) error {
	if len(m.ExpectedCalls) == 0 {
		return nil
	}
	args := m.Called(studentAddress, amount)
	return args.Error(0)
}

func (m *ContractAdapterMock) GetBalance(studentAddress common.Address) (*big.Int, error) {
	if len(m.ExpectedCalls) == 0 {
		return big.NewInt(0), nil
	}
	args := m.Called(studentAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}
