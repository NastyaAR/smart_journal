package adapters

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:generate mockgen -destination=mocks/contract_adapter_mock.go -package=mocks blockchain_project/internal/adapters ContractAdapter

type ContractAdapter interface {
	AwardTokens(studentAddress common.Address, amount *big.Int) error
	RedeemTokens(studentAddress common.Address, amount *big.Int) error
	GetBalance(studentAddress common.Address) (*big.Int, error)
}

type contractAdapter struct {
	client   *ethclient.Client
	address  common.Address
	contract *bind.BoundContract
	auth     *bind.TransactOpts
	txMu     sync.Mutex
}

const academicMeritTokenABI = `[
	{"inputs":[{"internalType":"address","name":"student","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"reason","type":"string"}],"name":"awardStudent","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"address","name":"student","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"redeemStudent","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"address","name":"student","type":"address"}],"name":"getBalance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

const defaultAwardReason = "Awarded from Smart Journal"

func NewContractAdapter(rpcURL, contractAddress string) (ContractAdapter, error) {
	if rpcURL == "" {
		return nil, errors.New("RPC_URL is required")
	}
	if contractAddress == "" {
		return nil, errors.New("CONTRACT_ADDRESS is required")
	}
	if !common.IsHexAddress(contractAddress) {
		return nil, fmt.Errorf("invalid CONTRACT_ADDRESS: %s", contractAddress)
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	contractABI, err := abi.JSON(strings.NewReader(academicMeritTokenABI))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	privateKeyHex := strings.TrimPrefix(os.Getenv("CONTRACT_ADMIN_PRIVATE_KEY"), "0x")
	if privateKeyHex == "" {
		client.Close()
		return nil, errors.New("CONTRACT_ADMIN_PRIVATE_KEY is required")
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to parse CONTRACT_ADMIN_PRIVATE_KEY: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to get chain id: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create transaction signer: %w", err)
	}

	address := common.HexToAddress(contractAddress)
	return &contractAdapter{
		client:   client,
		address:  address,
		contract: bind.NewBoundContract(address, contractABI, client, client, client),
		auth:     auth,
	}, nil
}

func (c *contractAdapter) AwardTokens(studentAddress common.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	return c.transactAndWait("awardStudent", studentAddress, amount, defaultAwardReason)
}

func (c *contractAdapter) RedeemTokens(studentAddress common.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	return c.transactAndWait("redeemStudent", studentAddress, amount)
}

func (c *contractAdapter) transactAndWait(method string, params ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	c.txMu.Lock()
	defer c.txMu.Unlock()

	auth := *c.auth
	auth.Context = ctx

	tx, err := c.contract.Transact(&auth, method, params...)
	if err != nil {
		return fmt.Errorf("failed to send %s transaction: %w", method, err)
	}

	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for %s transaction %s: %w", method, tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("%s transaction %s failed", method, tx.Hash().Hex())
	}

	return nil
}

func (c *contractAdapter) GetBalance(studentAddress common.Address) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result []any
	if err := c.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getBalance", studentAddress); err != nil {
		return nil, fmt.Errorf("failed to call getBalance: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.New("getBalance returned no values")
	}

	balance, ok := result[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected getBalance return type %T", result[0])
	}
	return balance, nil
}
