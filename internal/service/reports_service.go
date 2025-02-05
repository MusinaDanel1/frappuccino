package service

import (
	"frappuccino/internal/dal"
	"sort"
)

type ReportsService struct {
	orderRepo dal.OrderInterface
	menuRepo  dal.MenuInterface
}

func NewReportsService(orderRepo dal.OrderInterface, menuRepo dal.MenuInterface) *ReportsService {
	return &ReportsService{
		orderRepo: orderRepo,
		menuRepo:  menuRepo,
	}
}

func (s *ReportsService) TotalSales() (float64, error) {
	orders, err := s.orderRepo.List()
	if err != nil {
		return 0, nil
	}

	var totalSales float64

	for _, order := range orders {
		if order.Status == "closed" {
			for _, item := range order.Items {
				menuItem, err := s.menuRepo.GetByID(item.ProductID)
				if err != nil {
					return 0, err
				}
				totalSales += float64(item.Quantity) * menuItem.Price
			}
		}
	}
	return totalSales, nil
}

func (s *ReportsService) PopularItems() ([]string, error) {
	orders, err := s.orderRepo.List()
	if err != nil {
		return nil, err
	}
	popularity := make(map[string]int)
	for _, order := range orders {
		if order.Status != "closed" {
			continue
		}
		for _, item := range order.Items {
			popularity[item.ProductID] += item.Quantity
		}
	}

	type popularityItem struct {
		ProductID string
		Quantity  int
	}

	var popularityList []popularityItem
	for productID, quantity := range popularity {
		popularityList = append(popularityList, popularityItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	sort.Slice(popularityList, func(i, j int) bool {
		return popularityList[i].Quantity > popularityList[j].Quantity
	})

	var topItems []string
	for i := 0; i < 5 && i < len(popularityList); i++ {
		topItems = append(topItems, popularityList[i].ProductID)
	}

	return topItems, nil
}
