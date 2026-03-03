package llm

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/invopop/jsonschema"

	"restaurant-agent/pkg/tools"
)

type Client struct {
	API   *anthropic.Client
	Model anthropic.Model
}

func NewClient(apiKey string, model string) *Client {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		API:   &client,
		Model: anthropic.Model(model),
	}
}

// GenerateSchema creates a ToolInputSchemaParam from a Go struct using reflection.
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
		Required:   schema.Required,
	}
}

// BuildToolDefinitions creates Claude tool definitions matching the registry.
func BuildToolDefinitions() []anthropic.ToolUnionParam {
	return []anthropic.ToolUnionParam{
		// --- Inventory tools ---
		{OfTool: &anthropic.ToolParam{
			Name:        "check_inventory",
			Description: param.NewOpt("Check the current stock level of a specific ingredient. Returns quantity, unit, supplier, and reorder threshold."),
			InputSchema: GenerateSchema[tools.CheckInventoryInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "add_stock",
			Description: param.NewOpt("Add stock when a delivery arrives or inventory is received. Increases the quantity of an existing item or creates a new one."),
			InputSchema: GenerateSchema[tools.AddStockInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "remove_stock",
			Description: param.NewOpt("Remove stock due to waste, spoilage, or manual correction. Decreases the quantity of an item."),
			InputSchema: GenerateSchema[tools.RemoveStockInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "list_low_stock",
			Description: param.NewOpt("List all inventory items that are at or below their reorder threshold. Use this to identify what needs restocking."),
			InputSchema: GenerateSchema[struct{}](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "check_menu_feasibility",
			Description: param.NewOpt("Check if a specific menu item can be prepared for a given number of servings based on current inventory levels. Returns per-ingredient feasibility details."),
			InputSchema: GenerateSchema[tools.CheckMenuFeasibilityInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "place_order",
			Description: param.NewOpt("Place a restock purchase order for an ingredient. The order is sent to the item's default supplier."),
			InputSchema: GenerateSchema[tools.PlaceOrderInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "get_inventory_report",
			Description: param.NewOpt("Get a comprehensive inventory report showing all items, stock levels, and low-stock count. Optionally filter by category (produce, meat, dairy, seafood, pantry)."),
			InputSchema: GenerateSchema[tools.GetInventoryReportInput](),
		}},
		{OfTool: &anthropic.ToolParam{
			Name:        "get_menu_items",
			Description: param.NewOpt("List all menu items and their required ingredients per serving. Optionally filter by category."),
			InputSchema: GenerateSchema[tools.GetMenuItemsInput](),
		}},
	}
}

const SystemPrompt = `You are the AI operations assistant for De Gouden Lepel, a Dutch restaurant. You help managers run their restaurant efficiently by managing inventory.

## Your Capabilities

### Inventory Management
- Check and manage ingredient stock levels
- Track low-stock items that need reordering
- Record incoming deliveries (add stock) and waste/spoilage (remove stock)
- Check if menu items can be prepared based on current inventory
- Place restock orders with suppliers
- Generate inventory reports by category

## Guidelines
- Be concise and actionable. Restaurant managers are busy — get to the point.
- When you spot a problem, suggest a specific action (e.g., "I recommend ordering 10 kg of mozzarella from DairyDirect").
- Use tools to get accurate data. Never guess at numbers or inventory levels.
- When discussing menu items, use their Dutch names as they appear in the system.`
