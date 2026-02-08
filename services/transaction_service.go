package services

import (
	"context"
	"kasir-api/models"
	"kasir-api/repositories"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(items []models.CheckoutItem) (*models.Transaction, error) {
	return s.repo.CreateTransaction(items)
}

func (s *TransactionService) GetSummaryToday(ctx context.Context) (*models.SummaryToday, error) {
	return s.repo.GetSummaryToday(ctx)
}

func (s *TransactionService) GetBestSellerToday(ctx context.Context) (string, int, error) {
	return s.repo.GetBestSellerToday(ctx)
}
