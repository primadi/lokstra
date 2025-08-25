package salesflow

import (
	"github.com/primadi/lokstra/serviceapi"
)

type InventoryService interface {
	DecreaseStock(itemID string, qty int) error
}

type SalesSV struct {
	// Services (auto inject by lokstra)
	DBPool    serviceapi.DbPool `service:"lokstra.dbpool.main"`
	Inventory InventoryService  `service:"svc.inventory"`
	Logger    serviceapi.Logger `service:"logger"`
	I18n      serviceapi.I18n   `service:"i18n.default"`

	// Vars runtime
	SalesID    int64   `var:"sales_id"`
	TotalSales float64 `var:"total_sales"`

	// Schema for auth/tenant middleware
	Schema string `var:"schema"` // isi dari auth/tenant middleware
}
