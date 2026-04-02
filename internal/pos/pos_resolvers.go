package pos

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/graph"
	"github.com/greenpos/backend/internal/model"
)

// POSResolver handles GraphQL operations for POS terminals
type POSResolver struct {
	Service *POSService
}

// CreateQuickOrderInput represents the GraphQL input for quick order creation
type CreateQuickOrderInput struct {
	BranchID     uuid.UUID
	TableID      *uuid.UUID
	UserID       uuid.UUID
	CustomerName string
	Notes        string
	Items        []CreateQuickOrderItemInput
}

// CreateQuickOrderItemInput represents a single item in a quick order
type CreateQuickOrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	Notes     string
}

// CreateQuickOrderResult is the GraphQL result for quick order creation
type CreateQuickOrderResult struct {
	Order    *model.Order
	Items    []model.OrderItem
	Subtotal float64
	Tax      float64
	Total    float64
	Success  bool
	Message  string
}

// CloseTableResult is the GraphQL result for closing a table
type CloseTableResultGQL struct {
	Table        *graph.Table
	Orders       []model.Order
	Total        float64
	OrderCount   int
	TableFreed   bool
	Success      bool
	Message      string
}

// CreateQuickOrder is the GraphQL mutation for creating a quick order
func (r *POSResolver) CreateQuickOrder(ctx context.Context, input CreateQuickOrderInput) (*CreateQuickOrderResult, error) {
	items := make([]QuickOrderItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = QuickOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Notes:     item.Notes,
		}
	}

	posInput := QuickOrderInput{
		BranchID:     input.BranchID,
		TableID:      input.TableID,
		UserID:       input.UserID,
		CustomerName: input.CustomerName,
		Items:        items,
		Notes:        input.Notes,
	}

	result, err := r.Service.CreateQuickOrder(ctx, posInput)
	if err != nil {
		return &CreateQuickOrderResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &CreateQuickOrderResult{
		Order:    result.Order,
		Items:    result.Items,
		Subtotal: result.Subtotal,
		Tax:      result.Tax,
		Total:    result.Total,
		Success:  true,
		Message:  "Order created successfully",
	}, nil
}

// AddItemsToOrderInput represents the input for adding items to an order
type AddItemsToOrderInput struct {
	OrderID uuid.UUID
	Items   []CreateQuickOrderItemInput
}

// AddItemsResult is the GraphQL result for adding items to an order
type AddItemsResult struct {
	Order       *model.Order
	AddedItems  []model.OrderItem
	NewSubtotal float64
	NewTax      float64
	NewTotal    float64
	Success     bool
	Message     string
}

// AddItemsToOrder is the GraphQL mutation for adding items to an existing order
func (r *POSResolver) AddItemsToOrder(ctx context.Context, input AddItemsToOrderInput) (*AddItemsResult, error) {
	items := make([]QuickOrderItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = QuickOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Notes:     item.Notes,
		}
	}

	order, addedItems, err := r.Service.AddItemsToOrder(ctx, input.OrderID, items)
	if err != nil {
		return &AddItemsResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &AddItemsResult{
		Order:       order,
		AddedItems:  addedItems,
		NewSubtotal: order.Subtotal,
		NewTax:      order.Tax,
		NewTotal:    order.Total,
		Success:     true,
		Message:     "Items added successfully",
	}, nil
}

// CloseTableInput represents the input for closing a table
type CloseTableInput struct {
	TableID uuid.UUID
}

// CloseTable is the GraphQL mutation for closing a table and settling all orders
func (r *POSResolver) CloseTable(ctx context.Context, input CloseTableInput) (*CloseTableResultGQL, error) {
	result, err := r.Service.CloseTable(ctx, input.TableID)
	if err != nil {
		return &CloseTableResultGQL{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	orders := make([]model.Order, len(result.Orders))
	for i, o := range result.Orders {
		orders[i] = *o
	}

	// Convert table model to graph.Table type
	var tableGraph *graph.Table
	if result.Table != nil {
		t := graph.Table{
			ID:        result.Table.ID,
			BranchID:  result.Table.BranchID,
			Number:    result.Table.Number,
			Capacity:  result.Table.Capacity,
			Status:    result.Table.Status,
			CreatedAt: result.Table.CreatedAt,
			UpdatedAt: result.Table.UpdatedAt,
		}
		tableGraph = &t
	}

	return &CloseTableResultGQL{
		Table:      tableGraph,
		Orders:     orders,
		Total:      result.Total,
		OrderCount: result.OrderCount,
		TableFreed: result.TableFreed,
		Success:    true,
		Message:    "Table closed successfully",
	}, nil
}

// GetTableActiveOrders is a GraphQL query to get active orders for a table
type TableOrdersResult struct {
	Orders  []model.Order
	Success bool
	Message string
}

// GetTableActiveOrders returns all active orders for a table
func (r *POSResolver) GetTableActiveOrders(ctx context.Context, tableID uuid.UUID) (*TableOrdersResult, error) {
	orders, err := r.Service.GetTableOrders(ctx, tableID)
	if err != nil {
		return &TableOrdersResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	result := make([]model.Order, len(orders))
	for i, o := range orders {
		result[i] = *o
	}

	return &TableOrdersResult{
		Orders:  result,
		Success: true,
	}, nil
}
