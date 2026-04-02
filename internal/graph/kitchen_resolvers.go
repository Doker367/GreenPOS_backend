package graph

// Kitchen resolvers for the Kitchen Display System (KDS)
// These resolvers handle real-time order display for kitchen staff

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// KitchenOrders returns all orders in PREPARING status for a specific branch
// This is the main query for the Kitchen Display System
func (r *queryResolver) KitchenOrders(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	orders, err := r.Services.Order.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}

	kitchenOrders := make([]model.Order, 0)
	for _, o := range orders {
		if o.Status == model.OrderPreparing {
			kitchenOrders = append(kitchenOrders, o)
		}
	}
	return kitchenOrders, nil
}

// MarkOrderReady updates an order status to READY (kitchen completed preparation)
func (r *mutationResolver) MarkOrderReady(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	// Get order first for notification
	order, err := r.Services.Order.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.Services.Order.UpdateStatus(ctx, id, model.OrderReady)
	if err != nil {
		return nil, err
	}

	// Notify waiter that order is ready
	go func() {
		tableName := "Mostrador"
		if order.TableID != nil {
			if table, err := r.Services.Table.GetByID(context.Background(), *order.TableID); err == nil {
				tableName = table.Number
			}
		}
		r.Services.FCM.NotifyOrderReady(context.Background(), order.ID, order.UserID, tableName)
	}()

	return r.Services.Order.GetByID(ctx, id)
}

// KitchenOrderInfo provides additional context for kitchen display
type KitchenOrderInfo struct {
	// Time since order was created (in kitchen)
	WaitTimeMinutes int `json:"waitTimeMinutes"`
	// Is the order overdue (>10 minutes)
	IsOverdue bool `json:"isOverdue"`
	// Number of items in the order
	ItemCount int `json:"itemCount"`
	// Priority level based on wait time
	Priority string `json:"priority"` // "normal", "urgent", "critical"
}

// GetKitchenOrderInfo calculates additional info for kitchen display
func GetKitchenOrderInfo(order model.Order) KitchenOrderInfo {
	info := KitchenOrderInfo{
		ItemCount: len(order.Items),
		Priority:  "normal",
	}

	if !order.CreatedAt.IsZero() {
		waitDuration := time.Since(order.CreatedAt)
		info.WaitTimeMinutes = int(waitDuration.Minutes())

		// Critical: > 15 minutes
		if info.WaitTimeMinutes > 15 {
			info.Priority = "critical"
			info.IsOverdue = true
		} else if info.WaitTimeMinutes > 10 {
			// Urgent: > 10 minutes
			info.Priority = "urgent"
			info.IsOverdue = true
		} else if info.WaitTimeMinutes > 5 {
			// Warning: > 5 minutes
			info.Priority = "warning"
		}
	}

	return info
}

// Subscription support for real-time kitchen updates (optional enhancement)
// For now, we use polling every 3 seconds from the frontend

// OrderStatus changed event for subscriptions
type OrderStatusEvent struct {
	OrderID  uuid.UUID        `json:"orderId"`
	Status   model.OrderStatus `json:"status"`
	BranchID uuid.UUID        `json:"branchId"`
	Timestamp time.Time       `json:"timestamp"`
}