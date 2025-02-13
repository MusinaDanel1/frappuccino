package service

import (
	"context"
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type ReportService struct {
	repo *dal.ReportRepository
}

func NewReportService(repo *dal.ReportRepository) *ReportService {
	return &ReportService{
		repo: repo,
	}
}

func (s *ReportService) GetTotalSales(ctx context.Context) (float64, error) {
	totalSales, err := s.repo.GetTotalSales(ctx)
	if err != nil {
		return 0, fmt.Errorf("Could not get total sales: %w", err)
	}

	return totalSales, nil
}

func (s *ReportService) GetPopularItems(ctx context.Context) ([]string, error) {
	popularItems, err := s.repo.GetPopularItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not get popular items: %w", err)
	}

	return popularItems, nil
}

func (s *ReportService) GetOrderedItemsByPeriod(ctx context.Context, period string, month string, year string) (map[string]int, error) {
	if period != "day" && period != "month" {
		return nil, fmt.Errorf("invalid period: must be 'day' or 'month'")
	}

	orders, err := s.repo.GetOrderedItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get ordered items: %w", err)
	}

	groupedOrders := make(map[string]int)

	for _, order := range orders {
		if period == "day" {
			if month == "" {
				return nil, fmt.Errorf("month is required when period is 'day'")
			}
			orderMonth := order.CreatedAt.Month().String()
			if orderMonth == month {
				day := order.CreatedAt.Format("2006-01-02")
				groupedOrders[day]++
			}
		} else if period == "month" {
			if year == "" {
				return nil, fmt.Errorf("year is required when period is 'month'")
			}
			orderYear := order.CreatedAt.Format("2006")
			if orderYear == year {
				month := order.CreatedAt.Month().String()
				groupedOrders[month]++
			}
		}
	}

	return groupedOrders, nil
}

func (s *ReportService) Search(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (*models.SearchResponse, error) {
	return s.repo.FullTextSearch(ctx, query, filters, minPrice, maxPrice)
}
