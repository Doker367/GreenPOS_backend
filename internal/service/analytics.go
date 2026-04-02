package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// ============ ANALYTICS SERVICE ============

type AnalyticsService struct {
	orders   repository.OrderRepositoryInterface
	products repository.ProductRepositoryInterface
	tables   repository.TableRepositoryInterface
	log      *slog.Logger
}

func NewAnalyticsService(
	orders repository.OrderRepositoryInterface,
	products repository.ProductRepositoryInterface,
	tables repository.TableRepositoryInterface,
	log *slog.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		orders:   orders,
		products: products,
		tables:   tables,
		log:      log,
	}
}

// DashboardMetrics contains all key metrics for the dashboard
type DashboardMetrics struct {
	TotalRevenue    float64
	TodayRevenue    float64
	WeekRevenue     float64
	MonthRevenue    float64
	TotalOrders     int
	OrdersToday     int
	OrdersThisWeek  int
	OrdersThisMonth int
	AverageTicket   float64
	ActiveOrders    int
	TotalTables     int
	AvailableTables int
}

// DailySales represents sales data for a single day
type DailySales struct {
	Date     string  `json:"date"`
	Orders   int     `json:"orders"`
	Revenue  float64 `json:"revenue"`
}

// TopProductData represents a top selling product
type TopProductData struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	QuantitySold int    `json:"quantitySold"`
	Revenue     float64 `json:"revenue"`
}

// StatusCount represents order count by status
type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// RevenueReport represents a complete revenue report
type RevenueReport struct {
	Period          string   `json:"period"`
	TotalRevenue    float64  `json:"totalRevenue"`
	TotalOrders     int      `json:"totalOrders"`
	AverageTicket   float64  `json:"averageTicket"`
	ByStatus        []StatusCount `json:"byStatus"`
	DailySales      []DailySales  `json:"dailySales"`
}

// GetDashboardMetrics returns comprehensive dashboard metrics for a branch
func (s *AnalyticsService) GetDashboardMetrics(ctx context.Context, branchID uuid.UUID) (*DashboardMetrics, error) {
	orders, err := s.orders.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	tables, err := s.tables.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	var totalOrders, ordersToday, ordersThisWeek, ordersThisMonth int
	var totalRevenue, revenueToday, revenueThisWeek, revenueThisMonth float64
	var activeOrders int

	for _, order := range orders {
		isPaid := order.Status == model.OrderPaid || order.Status == model.OrderDelivered
		if !isPaid && order.Status != model.OrderCancelled {
			activeOrders++
		}
		if !isPaid {
			continue
		}

		totalOrders++
		totalRevenue += order.Total

		if order.CreatedAt.After(today) {
			ordersToday++
			revenueToday += order.Total
		}
		if order.CreatedAt.After(weekAgo) {
			ordersThisWeek++
			revenueThisWeek += order.Total
		}
		if order.CreatedAt.After(monthAgo) {
			ordersThisMonth++
			revenueThisMonth += order.Total
		}
	}

	avgTicket := float64(0)
	if totalOrders > 0 {
		avgTicket = totalRevenue / float64(totalOrders)
	}

	var availableTables int
	for _, t := range tables {
		if t.Status == model.TableAvailable {
			availableTables++
		}
	}

	return &DashboardMetrics{
		TotalRevenue:     totalRevenue,
		TodayRevenue:     revenueToday,
		WeekRevenue:      revenueThisWeek,
		MonthRevenue:    revenueThisMonth,
		TotalOrders:     totalOrders,
		OrdersToday:      ordersToday,
		OrdersThisWeek:   ordersThisWeek,
		OrdersThisMonth:  ordersThisMonth,
		AverageTicket:    avgTicket,
		ActiveOrders:     activeOrders,
		TotalTables:      len(tables),
		AvailableTables:  availableTables,
	}, nil
}

// GetSalesByDay returns daily sales for the specified number of days
func (s *AnalyticsService) GetSalesByDay(ctx context.Context, branchID uuid.UUID, days int) ([]DailySales, error) {
	orders, err := s.orders.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	now := time.Now()
	result := make([]DailySales, days)

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		var dayOrders int
		var dayRevenue float64
		for _, order := range orders {
			if (order.Status == model.OrderPaid || order.Status == model.OrderDelivered) &&
				order.CreatedAt.After(dayStart) && order.CreatedAt.Before(dayEnd) {
				dayOrders++
				dayRevenue += order.Total
			}
		}
		result[days-1-i] = DailySales{
			Date:    dayStart.Format("2006-01-02"),
			Orders:  dayOrders,
			Revenue: dayRevenue,
		}
	}
	return result, nil
}

// GetTopProducts returns the top selling products for a branch
func (s *AnalyticsService) GetTopProducts(ctx context.Context, branchID uuid.UUID, limit int) ([]TopProductData, error) {
	orders, err := s.orders.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	productSales := make(map[string]struct {
		name    string
		qty     int
		revenue float64
	})

	for _, order := range orders {
		if order.Status != model.OrderPaid && order.Status != model.OrderDelivered {
			continue
		}
		items, err := s.orders.GetItems(ctx, order.ID)
		if err != nil {
			continue
		}
		for _, item := range items {
			prod, _ := s.products.GetByID(ctx, item.ProductID)
			name := "Unknown"
			if prod != nil {
				name = prod.Name
			}
			key := item.ProductID.String()
			if existing, ok := productSales[key]; ok {
				productSales[key] = struct {
					name    string
					qty     int
					revenue float64
				}{name: name, qty: existing.qty + item.Quantity, revenue: existing.revenue + item.TotalPrice}
			} else {
				productSales[key] = struct {
					name    string
					qty     int
					revenue float64
				}{name: name, qty: item.Quantity, revenue: item.TotalPrice}
			}
		}
	}

	type productSale struct {
		id      string
		name    string
		qty     int
		revenue float64
	}
	var sorted []productSale
	for id, ps := range productSales {
		sorted = append(sorted, productSale{id: id, name: ps.name, qty: ps.qty, revenue: ps.revenue})
	}
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].revenue > sorted[i].revenue {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if limit > len(sorted) {
		limit = len(sorted)
	}

	var result []TopProductData
	for i := 0; i < limit; i++ {
		result = append(result, TopProductData{
			ProductID:    sorted[i].id,
			ProductName:  sorted[i].name,
			QuantitySold: sorted[i].qty,
			Revenue:      sorted[i].revenue,
		})
	}
	return result, nil
}

// GetOrdersByStatus returns order counts grouped by status
func (s *AnalyticsService) GetOrdersByStatus(ctx context.Context, branchID uuid.UUID) ([]StatusCount, error) {
	orders, err := s.orders.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	counts := make(map[string]int)
	for _, order := range orders {
		counts[string(order.Status)]++
	}

	var result []StatusCount
	for status, count := range counts {
		result = append(result, StatusCount{Status: status, Count: count})
	}
	return result, nil
}

// GetRevenueByPeriod returns revenue metrics for a specific period
func (s *AnalyticsService) GetRevenueByPeriod(ctx context.Context, branchID uuid.UUID, period string) (*RevenueReport, error) {
	orders, err := s.orders.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var startDate time.Time
	switch period {
	case "today":
		startDate = today
	case "week":
		startDate = today.AddDate(0, 0, -7)
	case "month":
		startDate = today.AddDate(0, -1, 0)
	case "year":
		startDate = today.AddDate(-1, 0, 0)
	default:
		startDate = time.Time{}
	}

	var totalOrders int
	var totalRevenue float64
	byStatus := make(map[string]int)
	dailyMap := make(map[string]float64)
	dailyOrdersMap := make(map[string]int)

	for _, order := range orders {
		if order.Status == model.OrderPaid || order.Status == model.OrderDelivered {
			if order.CreatedAt.After(startDate) {
				totalOrders++
				totalRevenue += order.Total
				byStatus[string(order.Status)]++
				
				day := order.CreatedAt.Format("2006-01-02")
				dailyMap[day] += order.Total
				dailyOrdersMap[day]++
			}
		}
	}

	avgTicket := float64(0)
	if totalOrders > 0 {
		avgTicket = totalRevenue / float64(totalOrders)
	}

	var statusResult []StatusCount
	for status, count := range byStatus {
		statusResult = append(statusResult, StatusCount{Status: status, Count: count})
	}

	var dailySales []DailySales
	for day, rev := range dailyMap {
		dailySales = append(dailySales, DailySales{
			Date:    day,
			Revenue: rev,
			Orders:  dailyOrdersMap[day],
		})
	}

	return &RevenueReport{
		Period:         period,
		TotalRevenue:   totalRevenue,
		TotalOrders:    totalOrders,
		AverageTicket:  avgTicket,
		ByStatus:       statusResult,
		DailySales:     dailySales,
	}, nil
}
