package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

func (ec *executionContext) unmarshalInputUUID(ctx context.Context, v interface{}) (uuid.UUID, error) {
	switch val := v.(type) {
	case string:
		return uuid.Parse(val)
	case []byte:
		return uuid.Parse(string(val))
	default:
		return uuid.Nil, fmt.Errorf("cannot unmarshal UUID from %T", v)
	}
}

func (ec *executionContext) _UUID(ctx context.Context, sel interface{}, v *uuid.UUID) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	return graphql.MarshalString(v.String())
}

func (ec *executionContext) unmarshalInputJSON(ctx context.Context, v interface{}) (model.JSON, error) {
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

func (ec *executionContext) _JSON(ctx context.Context, sel interface{}, v model.JSON) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	b, err := json.Marshal(v)
	if err != nil {
		return graphql.Null
	}
	return graphql.MarshalString(string(b))
}
