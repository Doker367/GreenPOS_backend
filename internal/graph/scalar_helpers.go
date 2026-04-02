package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// UUID marshaling helpers

func unmarshalInputUUID(ctx context.Context, v interface{}) (uuid.UUID, error) {
	switch val := v.(type) {
	case string:
		return uuid.Parse(val)
	case []byte:
		return uuid.Parse(string(val))
	default:
		return uuid.Nil, fmt.Errorf("cannot unmarshal UUID from %T", v)
	}
}

func marshalUUID(v *uuid.UUID) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	return graphql.MarshalString(v.String())
}

func unmarshalUUID(ctx context.Context, v interface{}) (*uuid.UUID, error) {
	if v == nil {
		return nil, nil
	}
	u, err := unmarshalInputUUID(ctx, v)
	return &u, err
}

// JSON marshaling helpers

func unmarshalInputJSON(ctx context.Context, v interface{}) (model.JSON, error) {
	switch val := v.(type) {
	case string:
		if val == "" {
			return nil, nil
		}
		var result model.JSON
		err := json.Unmarshal([]byte(val), &result)
		return result, err
	case []byte:
		if len(val) == 0 {
			return nil, nil
		}
		var result model.JSON
		err := json.Unmarshal(val, &result)
		return result, err
	default:
		return nil, fmt.Errorf("cannot unmarshal JSON from %T", v)
	}
}

func marshalJSON(v model.JSON) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	b, err := json.Marshal(v)
	if err != nil {
		return graphql.Null
	}
	return graphql.MarshalString(string(b))
}

// Time marshaling

func marshalDateTime(v *time.Time) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	return graphql.MarshalTime(*v)
}

func unmarshalDateTime(ctx context.Context, v interface{}) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case time.Time:
		return &val, nil
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return &t, nil
	default:
		return nil, fmt.Errorf("cannot unmarshal time from %T", v)
	}
}

// Enum marshaling helpers

func marshalOrderStatus(v model.OrderStatus) graphql.Marshaler {
	return graphql.MarshalString(string(v))
}

func unmarshalOrderStatus(ctx context.Context, v interface{}) (model.OrderStatus, error) {
	switch val := v.(type) {
	case string:
		return model.OrderStatus(val), nil
	default:
		return "", fmt.Errorf("cannot unmarshal OrderStatus from %T", v)
	}
}

func marshalTableStatus(v model.TableStatus) graphql.Marshaler {
	return graphql.MarshalString(string(v))
}

func unmarshalTableStatus(ctx context.Context, v interface{}) (model.TableStatus, error) {
	switch val := v.(type) {
	case string:
		return model.TableStatus(val), nil
	default:
		return "", fmt.Errorf("cannot unmarshal TableStatus from %T", v)
	}
}

func marshalUserRole(v model.UserRole) graphql.Marshaler {
	return graphql.MarshalString(string(v))
}

func unmarshalUserRole(ctx context.Context, v interface{}) (model.UserRole, error) {
	switch val := v.(type) {
	case string:
		return model.UserRole(val), nil
	default:
		return "", fmt.Errorf("cannot unmarshal UserRole from %T", v)
	}
}

func marshalReservationStatus(v model.ReservationStatus) graphql.Marshaler {
	return graphql.MarshalString(string(v))
}

func unmarshalReservationStatus(ctx context.Context, v interface{}) (model.ReservationStatus, error) {
	switch val := v.(type) {
	case string:
		return model.ReservationStatus(val), nil
	default:
		return "", fmt.Errorf("cannot unmarshal ReservationStatus from %T", v)
	}
}

func marshalInvoiceStatus(v model.InvoiceStatus) graphql.Marshaler {
	return graphql.MarshalString(string(v))
}

func unmarshalInvoiceStatus(ctx context.Context, v interface{}) (model.InvoiceStatus, error) {
	switch val := v.(type) {
	case string:
		return model.InvoiceStatus(val), nil
	default:
		return "", fmt.Errorf("cannot unmarshal InvoiceStatus from %T", v)
	}
}

// Scalar marshaling for GraphQL ID type (maps to UUID)
func marshalID(v *uuid.UUID) graphql.Marshaler {
	return marshalUUID(v)
}

func unmarshalID(ctx context.Context, v interface{}) (*uuid.UUID, error) {
	return unmarshalUUID(ctx, v)
}
