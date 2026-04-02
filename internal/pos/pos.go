package pos

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// TaxRate is the default tax rate (16% IVA in Mexico)
const TaxRate = 0.16

var (
	ErrOrderNotFound     = fmt.Errorf("order not found")
	ErrTableNotFound     = fmt.Errorf("table not found")
	ErrProductNotFound   = fmt.Errorf("product not found")
	ErrProductUnavailable = fmt.Errorf("product not available")
	ErrOrderNotModifiable = fmt.Errorf("order cannot be modified in current status")
)

// POSService handles Point of Sale operations
// Supports fast order entry, multi-order per table, and table closing
type POSService struct {
	orders   repository.OrderRepositoryInterface
	products repository.ProductRepositoryInterface
	tables   repository.TableRepositoryInterface
	taxRate  float64
}

// NewPOSService creates a new POS service
func NewPOSService(
	orders repository.OrderRepositoryInterface,
	products repository.ProductRepositoryInterface,
	tables repository.TableRepositoryInterface,
) *POSService {
	return &POSService{
		orders:   orders,
		products: products,
		tables:   tables,
		taxRate:  TaxRate,
	}
}

// QuickOrderInput represents a quick order with items for POS
type QuickOrderInput struct {
	BranchID     uuid.UUID
	TableID      *uuid.UUID // nil for takeout/delivery
	UserID       uuid.UUID
	CustomerName string
	Items        []QuickOrderItem
	Notes        string
}

// QuickOrderItem is a simplified order item for POS
type QuickOrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	Notes     string
}

// QuickOrderResult is the result of creating a quick order
type QuickOrderResult struct {
	Order    *model.Order
	Items    []model.OrderItem
	Subtotal float64
	Tax      float64
	Total    float64
}

// CreateQuickOrder creates a new order with items in one step
// This is optimized for POS terminals where speed is critical
func (s *POSService) CreateQuickOrder(ctx context.Context, input QuickOrderInput) (*QuickOrderResult, error) {
	// Validate table if provided
	if input.TableID != nil {
		table, err := s.tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		// Note: We allow multiple orders per table, so we don't check for existing orders
		// The table status will be updated when ALL orders are closed
		if table.Status != model.TableAvailable && table.Status != model.TableOccupied {
			return nil, fmt.Errorf("table is not available: %s", table.Status)
		}
	}

	// Validate all products and calculate totals
	var subtotal float64
	orderItems := make([]model.OrderItem, 0, len(input.Items))

	for _, item := range input.Items {
		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrProductNotFound, item.ProductID)
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("%w: %s", ErrProductUnavailable, product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		subtotal += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			ID:         uuid.New(),
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
			Notes:      item.Notes,
		})
	}

	tax := subtotal * s.taxRate
	total := subtotal + tax

	// Create the order
	order := &model.Order{
		ID:            uuid.New(),
		BranchID:      input.BranchID,
		TableID:       input.TableID,
		UserID:        input.UserID,
		CustomerName:  input.CustomerName,
		Status:        model.OrderPending,
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         total,
		Notes:         input.Notes,
	}

	if err := s.orders.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Add items to the order
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := s.orders.AddItem(ctx, &orderItems[i]); err != nil {
			return nil, fmt.Errorf("failed to add item: %w", err)
		}
	}

	// Update table status to occupied if this is a table order
	if input.TableID != nil {
		_ = s.tables.UpdateStatus(ctx, *input.TableID, model.TableOccupied)
	}

	return &QuickOrderResult{
		Order:    order,
		Items:    orderItems,
		Subtotal: subtotal,
		Tax:      tax,
		Total:    total,
	}, nil
}

// AddItemsToOrder adds items to an existing order
func (s *POSService) AddItemsToOrder(ctx context.Context, orderID uuid.UUID, items []QuickOrderItem) (*model.Order, []model.OrderItem, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, nil, ErrOrderNotFound
	}

	// Only allow adding items to pending or accepted orders
	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return nil, nil, ErrOrderNotModifiable
	}

	var addedItems []model.OrderItem
	var additionalSubtotal float64

	for _, item := range items {
		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s", ErrProductNotFound, item.ProductID)
		}
		if !product.IsAvailable {
			return nil, nil, fmt.Errorf("%w: %s", ErrProductUnavailable, product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		additionalSubtotal += itemTotal

		orderItem := &model.OrderItem{
			ID:         uuid.New(),
			OrderID:    orderID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
			Notes:      item.Notes,
		}

		if err := s.orders.AddItem(ctx, orderItem); err != nil {
			return nil, nil, fmt.Errorf("failed to add item: %w", err)
		}
		addedItems = append(addedItems, *orderItem)
	}

	// Recalculate order totals
	newSubtotal := order.Subtotal + additionalSubtotal
	newTax := newSubtotal * s.taxRate
	newTotal := newSubtotal + newTax - order.Discount

	order.Subtotal = newSubtotal
	order.Tax = newTax
	order.Total = newTotal

	if err := s.orders.Update(ctx, order); err != nil {
		return nil, nil, fmt.Errorf("failed to update order totals: %w", err)
	}

	return order, addedItems, nil
}

// CloseTable closes all active orders for a table and frees the table
// Returns the total of all orders combined for final payment
func (s *POSService) CloseTable(ctx context.Context, tableID uuid.UUID) (*CloseTableResult, error) {
	table, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, ErrTableNotFound
	}

	// Get all active orders for this table
	// Note: We need to get all non-terminal orders
	allOrders, err := s.orders.GetByBranch(ctx, table.BranchID)
	if err != nil {
		return nil, err
	}

	var activeOrders []*model.Order
	for _, o := range allOrders {
		if o.TableID != nil && *o.TableID == tableID {
			if o.Status != model.OrderPaid && o.Status != model.OrderCancelled && o.Status != model.OrderDelivered {
				activeOrders = append(activeOrders, &o)
			}
		}
	}

	if len(activeOrders) == 0 {
		// No active orders, just free the table
		if err := s.tables.UpdateStatus(ctx, tableID, model.TableAvailable); err != nil {
			return nil, err
		}
		return &CloseTableResult{
			Table:            table,
			Orders:           []*model.Order{},
			Total:            0,
			OrderCount:       0,
			TableFreed:       true,
		}, nil
	}

	// Mark all orders as delivered (ready for payment)
	grandTotal := 0.0
	for _, order := range activeOrders {
		if order.Status == model.OrderPending || order.Status == model.OrderAccepted || order.Status == model.OrderPreparing || order.Status == model.OrderReady {
			if err := s.orders.UpdateStatus(ctx, order.ID, model.OrderDelivered); err != nil {
				return nil, fmt.Errorf("failed to update order %s: %w", order.ID, err)
			}
			order.Status = model.OrderDelivered
		}
		grandTotal += order.Total
	}

	// Free the table
	if err := s.tables.UpdateStatus(ctx, tableID, model.TableAvailable); err != nil {
		return nil, fmt.Errorf("failed to free table: %w", err)
	}

	return &CloseTableResult{
		Table:        table,
		Orders:       activeOrders,
		Total:        grandTotal,
		OrderCount:   len(activeOrders),
		TableFreed:   true,
	}, nil
}

// GetTableOrders returns all active orders for a table
func (s *POSService) GetTableOrders(ctx context.Context, tableID uuid.UUID) ([]*model.Order, error) {
	allOrders, err := s.orders.GetByBranch(ctx, uuid.Nil) // Will need branch context
	if err != nil {
		return nil, err
	}

	// This would need branch context - for now we'll use a simplified approach
	_ = allOrders

	// Get the table first to know the branch
	table, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, ErrTableNotFound
	}

	orders, err := s.orders.GetByBranch(ctx, table.BranchID)
	if err != nil {
		return nil, err
	}

	var result []*model.Order
	for _, o := range orders {
		if o.TableID != nil && *o.TableID == tableID {
			if o.Status != model.OrderPaid && o.Status != model.OrderCancelled && o.Status != model.OrderDelivered {
				result = append(result, &o)
			}
		}
	}
	return result, nil
}

// CloseTableResult contains the results of closing a table
type CloseTableResult struct {
	Table        *model.RestaurantTable
	Orders       []*model.Order
	Total        float64
	OrderCount   int
	TableFreed   bool
}
