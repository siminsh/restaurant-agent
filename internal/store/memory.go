package store

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type InventoryItem struct {
	Name             string  `json:"name"`
	Quantity         float64 `json:"quantity"`
	Unit             string  `json:"unit"`
	ReorderThreshold float64 `json:"reorder_threshold"`
	Category         string  `json:"category"`
	Supplier         string  `json:"supplier"`
	LastUpdated      string  `json:"last_updated"`
}

type MenuItem struct {
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	Ingredients map[string]Ingredient `json:"ingredients"`
}

type Ingredient struct {
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type Order struct {
	ID        string  `json:"id"`
	ItemName  string  `json:"item_name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	Supplier  string  `json:"supplier"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type MemoryStore struct {
	mu        sync.RWMutex
	inventory map[string]*InventoryItem
	menu      map[string]*MenuItem
	orders    []Order
	orderSeq  int
}

func New() *MemoryStore {
	s := &MemoryStore{
		inventory: make(map[string]*InventoryItem),
		menu:      make(map[string]*MenuItem),
		orders:    []Order{},
	}
	s.seed()
	return s
}

func (s *MemoryStore) seed() {
	// Inventory items for the restaurant
	items := []InventoryItem{
		// Produce
		{Name: "aardappelen", Quantity: 25, Unit: "kg", ReorderThreshold: 15, Category: "produce", Supplier: "Groenteboer van Dam"},
		{Name: "uien", Quantity: 8, Unit: "kg", ReorderThreshold: 5, Category: "produce", Supplier: "Groenteboer van Dam"},
		{Name: "boerenkool", Quantity: 4, Unit: "kg", ReorderThreshold: 6, Category: "produce", Supplier: "Groenteboer van Dam"},
		{Name: "tomaten", Quantity: 3, Unit: "kg", ReorderThreshold: 5, Category: "produce", Supplier: "Groenteboer van Dam"},
		{Name: "sla", Quantity: 2, Unit: "kg", ReorderThreshold: 3, Category: "produce", Supplier: "Groenteboer van Dam"},
		{Name: "pompoen", Quantity: 5, Unit: "kg", ReorderThreshold: 4, Category: "produce", Supplier: "Groenteboer van Dam"},
		// Meat
		{Name: "biefstuk", Quantity: 6, Unit: "kg", ReorderThreshold: 8, Category: "meat", Supplier: "Slagerij de Wit"},
		{Name: "ossenhaas", Quantity: 3, Unit: "kg", ReorderThreshold: 4, Category: "meat", Supplier: "Slagerij de Wit"},
		{Name: "kippenpoten", Quantity: 5, Unit: "kg", ReorderThreshold: 6, Category: "meat", Supplier: "Slagerij de Wit"},
		{Name: "lamsbout", Quantity: 4, Unit: "kg", ReorderThreshold: 3, Category: "meat", Supplier: "Slagerij de Wit"},
		{Name: "eendenborst", Quantity: 2, Unit: "kg", ReorderThreshold: 3, Category: "meat", Supplier: "Slagerij de Wit"},
		// Seafood
		{Name: "zalm", Quantity: 5, Unit: "kg", ReorderThreshold: 4, Category: "seafood", Supplier: "Vishandel Noordzee"},
		{Name: "mosselen", Quantity: 8, Unit: "kg", ReorderThreshold: 10, Category: "seafood", Supplier: "Vishandel Noordzee"},
		{Name: "schol", Quantity: 3, Unit: "kg", ReorderThreshold: 4, Category: "seafood", Supplier: "Vishandel Noordzee"},
		{Name: "garnalen", Quantity: 2, Unit: "kg", ReorderThreshold: 3, Category: "seafood", Supplier: "Vishandel Noordzee"},
		// Dairy
		{Name: "boter", Quantity: 4, Unit: "kg", ReorderThreshold: 3, Category: "dairy", Supplier: "Zuivelhoeve"},
		{Name: "room", Quantity: 5, Unit: "liters", ReorderThreshold: 4, Category: "dairy", Supplier: "Zuivelhoeve"},
		{Name: "kaas", Quantity: 3, Unit: "kg", ReorderThreshold: 4, Category: "dairy", Supplier: "Zuivelhoeve"},
		{Name: "eieren", Quantity: 60, Unit: "stuks", ReorderThreshold: 30, Category: "dairy", Supplier: "Zuivelhoeve"},
		// Pantry
		{Name: "meel", Quantity: 12, Unit: "kg", ReorderThreshold: 8, Category: "pantry", Supplier: "Groothandel Bakker"},
		{Name: "rijst", Quantity: 10, Unit: "kg", ReorderThreshold: 6, Category: "pantry", Supplier: "Groothandel Bakker"},
		{Name: "pasta", Quantity: 8, Unit: "kg", ReorderThreshold: 5, Category: "pantry", Supplier: "Groothandel Bakker"},
		{Name: "olijfolie", Quantity: 6, Unit: "liters", ReorderThreshold: 3, Category: "pantry", Supplier: "Mediterrane Import"},
		{Name: "erwten", Quantity: 3, Unit: "kg", ReorderThreshold: 5, Category: "pantry", Supplier: "Groothandel Bakker"},
		{Name: "rookworst", Quantity: 4, Unit: "kg", ReorderThreshold: 5, Category: "pantry", Supplier: "Slagerij de Wit"},
	}

	now := time.Now().Format(time.RFC3339)
	for i := range items {
		items[i].LastUpdated = now
		s.inventory[items[i].Name] = &items[i]
	}

	// Menu items
	s.menu["stamppot boerenkool"] = &MenuItem{
		Name:     "Stamppot Boerenkool",
		Category: "main",
		Ingredients: map[string]Ingredient{
			"aardappelen": {Quantity: 0.3, Unit: "kg"},
			"boerenkool":  {Quantity: 0.15, Unit: "kg"},
			"rookworst":   {Quantity: 0.15, Unit: "kg"},
			"boter":       {Quantity: 0.02, Unit: "kg"},
			"uien":        {Quantity: 0.05, Unit: "kg"},
		},
	}
	s.menu["biefstuk"] = &MenuItem{
		Name:     "Biefstuk",
		Category: "main",
		Ingredients: map[string]Ingredient{
			"biefstuk":    {Quantity: 0.25, Unit: "kg"},
			"aardappelen": {Quantity: 0.2, Unit: "kg"},
			"boter":       {Quantity: 0.03, Unit: "kg"},
			"sla":         {Quantity: 0.05, Unit: "kg"},
		},
	}
	s.menu["ossenhaas"] = &MenuItem{
		Name:     "Ossenhaas",
		Category: "main",
		Ingredients: map[string]Ingredient{
			"ossenhaas": {Quantity: 0.2, Unit: "kg"},
			"boter":     {Quantity: 0.03, Unit: "kg"},
			"room":      {Quantity: 0.05, Unit: "liters"},
		},
	}
	s.menu["erwtensoep"] = &MenuItem{
		Name:     "Erwtensoep",
		Category: "soup",
		Ingredients: map[string]Ingredient{
			"erwten":    {Quantity: 0.15, Unit: "kg"},
			"rookworst": {Quantity: 0.1, Unit: "kg"},
			"uien":      {Quantity: 0.05, Unit: "kg"},
			"aardappelen": {Quantity: 0.1, Unit: "kg"},
		},
	}
	s.menu["zalm filet"] = &MenuItem{
		Name:     "Zalm Filet",
		Category: "fish",
		Ingredients: map[string]Ingredient{
			"zalm":       {Quantity: 0.2, Unit: "kg"},
			"olijfolie":  {Quantity: 0.02, Unit: "liters"},
			"aardappelen": {Quantity: 0.15, Unit: "kg"},
			"sla":        {Quantity: 0.05, Unit: "kg"},
		},
	}
	s.menu["mosselen"] = &MenuItem{
		Name:     "Mosselen (per kilo)",
		Category: "fish",
		Ingredients: map[string]Ingredient{
			"mosselen": {Quantity: 1.0, Unit: "kg"},
			"uien":     {Quantity: 0.05, Unit: "kg"},
			"room":     {Quantity: 0.1, Unit: "liters"},
			"boter":    {Quantity: 0.02, Unit: "kg"},
		},
	}
	s.menu["bitterballen"] = &MenuItem{
		Name:     "Bitterballen (8 stuks)",
		Category: "starter",
		Ingredients: map[string]Ingredient{
			"meel":  {Quantity: 0.05, Unit: "kg"},
			"boter": {Quantity: 0.03, Unit: "kg"},
			"eieren": {Quantity: 1, Unit: "stuks"},
		},
	}
	s.menu["tomatensoep"] = &MenuItem{
		Name:     "Tomatensoep",
		Category: "soup",
		Ingredients: map[string]Ingredient{
			"tomaten": {Quantity: 0.3, Unit: "kg"},
			"uien":    {Quantity: 0.05, Unit: "kg"},
			"room":    {Quantity: 0.05, Unit: "liters"},
		},
	}
}

func (s *MemoryStore) GetItem(name string) (InventoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.inventory[strings.ToLower(name)]
	if !ok {
		return InventoryItem{}, fmt.Errorf("item %q not found in inventory", name)
	}
	return *item, nil
}

func (s *MemoryStore) ListAll() []InventoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]InventoryItem, 0, len(s.inventory))
	for _, item := range s.inventory {
		items = append(items, *item)
	}
	return items
}

func (s *MemoryStore) AddStock(name string, quantity float64, unit string) (InventoryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(name)
	item, ok := s.inventory[key]
	if !ok {
		item = &InventoryItem{
			Name:             key,
			Quantity:         0,
			Unit:             unit,
			ReorderThreshold: quantity * 0.5,
			Category:         "other",
			Supplier:         "unknown",
		}
		s.inventory[key] = item
	}

	item.Quantity += quantity
	item.LastUpdated = time.Now().Format(time.RFC3339)
	return *item, nil
}

func (s *MemoryStore) RemoveStock(name string, quantity float64, reason string) (InventoryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(name)
	item, ok := s.inventory[key]
	if !ok {
		return InventoryItem{}, fmt.Errorf("item %q not found in inventory", name)
	}

	if item.Quantity < quantity {
		return InventoryItem{}, fmt.Errorf("insufficient stock: have %.1f %s of %s, trying to remove %.1f", item.Quantity, item.Unit, name, quantity)
	}

	item.Quantity -= quantity
	item.LastUpdated = time.Now().Format(time.RFC3339)
	return *item, nil
}

func (s *MemoryStore) GetLowStock() []InventoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var low []InventoryItem
	for _, item := range s.inventory {
		if item.Quantity <= item.ReorderThreshold {
			low = append(low, *item)
		}
	}
	return low
}

func (s *MemoryStore) CheckMenuFeasibility(menuItem string, servings int) (bool, map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := strings.ToLower(menuItem)
	mi, ok := s.menu[key]
	if !ok {
		return false, nil, fmt.Errorf("menu item %q not found", menuItem)
	}

	feasible := true
	details := make(map[string]string)
	for ingName, ing := range mi.Ingredients {
		needed := ing.Quantity * float64(servings)
		inv, ok := s.inventory[ingName]
		if !ok {
			feasible = false
			details[ingName] = fmt.Sprintf("NOT IN STOCK (need %.2f %s)", needed, ing.Unit)
			continue
		}
		if inv.Quantity < needed {
			feasible = false
			details[ingName] = fmt.Sprintf("INSUFFICIENT: have %.2f %s, need %.2f %s", inv.Quantity, inv.Unit, needed, ing.Unit)
		} else {
			details[ingName] = fmt.Sprintf("OK: have %.2f %s, need %.2f %s", inv.Quantity, inv.Unit, needed, ing.Unit)
		}
	}
	return feasible, details, nil
}

func (s *MemoryStore) PlaceOrder(itemName string, quantity float64) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(itemName)
	item, ok := s.inventory[key]
	supplier := "General Supplier"
	unit := "units"
	if ok {
		supplier = item.Supplier
		unit = item.Unit
	}

	s.orderSeq++
	order := Order{
		ID:        fmt.Sprintf("ORD-%04d", s.orderSeq),
		ItemName:  key,
		Quantity:  quantity,
		Unit:      unit,
		Supplier:  supplier,
		Status:    "placed",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.orders = append(s.orders, order)
	return order, nil
}

func (s *MemoryStore) GetInventoryReport(category string) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []InventoryItem
	var lowCount int
	for _, item := range s.inventory {
		if category != "" && !strings.EqualFold(item.Category, category) {
			continue
		}
		items = append(items, *item)
		if item.Quantity <= item.ReorderThreshold {
			lowCount++
		}
	}

	return map[string]interface{}{
		"total_items":     len(items),
		"low_stock_count": lowCount,
		"items":           items,
		"pending_orders":  len(s.orders),
	}
}

func (s *MemoryStore) GetMenuItems(category string) []MenuItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []MenuItem
	for _, mi := range s.menu {
		if category != "" && !strings.EqualFold(mi.Category, category) {
			continue
		}
		items = append(items, *mi)
	}
	return items
}
