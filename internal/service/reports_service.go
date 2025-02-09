package service

import (
	"context"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type ReportsService struct {
	reportsRepo dal.ReportInterface
}

func NewReportsService(reportsRepo dal.ReportInterface) *ReportsService {
	return &ReportsService{
		reportsRepo: reportsRepo,
	}
}

func (s *ReportsService) Search(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (*models.SearchResponse, error) {
	return s.reportsRepo.FullTextSearch(ctx, query, filters, minPrice, maxPrice)
}
