package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"restaurant-agent/internal/store"
)

// Input structs for each tool

type CheckInventoryInput struct {
	ItemName string `json:"item_name" jsonschema:"required,description=Name of the ingredient to check"`
}

type AddStockInput struct {
	ItemName string  `json:"item_name" jsonschema:"required,description=Name of the ingredient"`
	Quantity float64 `json:"quantity" jsonschema:"required,description=Amount to add"`
	Unit     string  `json:"unit" jsonschema:"required,description=Unit of measurement (kg or liters or units)"`
}

type RemoveStockInput struct {
	ItemName string  `json:"item_name" jsonschema:"required,description=Name of the ingredient"`
	Quantity float64 `json:"quantity" jsonschema:"required,description=Amount to remove"`
	Reason   string  `json:"reason" jsonschema:"required,description=Reason for removal (waste or spoilage or usage or correction)"`
}

type CheckMenuFeasibilityInput struct {
	MenuItem string `json:"menu_item" jsonschema:"required,description=Name of the menu item to check"`
	Servings int    `json:"servings" jsonschema:"required,description=Number of servings to check feasibility for"`
}

type PlaceOrderInput struct {
	ItemName string  `json:"item_name" jsonschema:"required,description=Name of the ingredient to order"`
	Quantity float64 `json:"quantity" jsonschema:"required,description=Amount to order"`
}

type GetInventoryReportInput struct {
	Category string `json:"category,omitempty" jsonschema:"description=Optional category filter (produce or meat or dairy or seafood or pantry)"`
}

type GetMenuItemsInput struct {
	Category string `json:"category,omitempty" jsonschema:"description=Optional category filter (pizza or pasta or main or appetizer)"`
}

// RegisterInventoryTools registers all inventory-related tools with the registry.
func RegisterInventoryTools(reg *Registry, s *store.MemoryStore) {
	reg.Register("check_inventory", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in CheckInventoryInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		item, err := s.GetItem(in.ItemName)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}
		result, _ := json.Marshal(item)
		return string(result), nil
	})

	reg.Register("add_stock", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in AddStockInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		if in.Quantity <= 0 {
			return fmt.Sprintf("Error: quantity must be positive, got %.2f", in.Quantity), nil
		}
		item, err := s.AddStock(in.ItemName, in.Quantity, in.Unit)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}
		result, _ := json.Marshal(map[string]interface{}{
			"message": fmt.Sprintf("Added %.1f %s of %s", in.Quantity, in.Unit, in.ItemName),
			"item":    item,
		})
		return string(result), nil
	})

	reg.Register("remove_stock", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in RemoveStockInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		if in.Quantity <= 0 {
			return fmt.Sprintf("Error: quantity must be positive, got %.2f", in.Quantity), nil
		}
		item, err := s.RemoveStock(in.ItemName, in.Quantity, in.Reason)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}
		result, _ := json.Marshal(map[string]interface{}{
			"message": fmt.Sprintf("Removed %.1f %s of %s (reason: %s)", in.Quantity, item.Unit, in.ItemName, in.Reason),
			"item":    item,
		})
		return string(result), nil
	})

	reg.Register("list_low_stock", func(ctx context.Context, input json.RawMessage) (string, error) {
		items := s.GetLowStock()
		if len(items) == 0 {
			return `{"message": "All items are above reorder thresholds. Inventory is in good shape!"}`, nil
		}
		result, _ := json.Marshal(map[string]interface{}{
			"low_stock_items": items,
			"count":           len(items),
		})
		return string(result), nil
	})

	reg.Register("check_menu_feasibility", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in CheckMenuFeasibilityInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		if in.Servings <= 0 {
			return fmt.Sprintf("Error: servings must be positive, got %d", in.Servings), nil
		}
		feasible, details, err := s.CheckMenuFeasibility(in.MenuItem, in.Servings)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}
		result, _ := json.Marshal(map[string]interface{}{
			"menu_item":  in.MenuItem,
			"servings":   in.Servings,
			"feasible":   feasible,
			"details":    details,
		})
		return string(result), nil
	})

	reg.Register("place_order", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in PlaceOrderInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		if in.Quantity <= 0 {
			return fmt.Sprintf("Error: quantity must be positive, got %.2f", in.Quantity), nil
		}
		order, err := s.PlaceOrder(in.ItemName, in.Quantity)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error()), nil
		}
		result, _ := json.Marshal(map[string]interface{}{
			"message": fmt.Sprintf("Order %s placed for %.1f %s of %s from %s", order.ID, order.Quantity, order.Unit, order.ItemName, order.Supplier),
			"order":   order,
		})
		return string(result), nil
	})

	reg.Register("get_inventory_report", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in GetInventoryReportInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		report := s.GetInventoryReport(in.Category)
		result, _ := json.Marshal(report)
		return string(result), nil
	})

	reg.Register("get_menu_items", func(ctx context.Context, input json.RawMessage) (string, error) {
		var in GetMenuItemsInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", fmt.Errorf("invalid input: %w", err)
		}
		items := s.GetMenuItems(in.Category)
		result, _ := json.Marshal(map[string]interface{}{
			"menu_items": items,
			"count":      len(items),
		})
		return string(result), nil
	})
}
