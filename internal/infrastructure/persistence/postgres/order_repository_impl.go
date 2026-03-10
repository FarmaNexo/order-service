package postgres

import (
	"context"
	"fmt"

	"github.com/farmanexo/order-service/internal/domain/entities"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewOrderRepository(db *gorm.DB, logger *zap.Logger) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{db: db, logger: logger}
}

func (r *OrderRepositoryImpl) Create(ctx context.Context, order *entities.Order) error {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepositoryImpl) FindByID(ctx context.Context, id string) (*entities.Order, error) {
	var order entities.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) FindByIDWithItems(ctx context.Context, id string) (*entities.Order, error) {
	var order entities.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) FindByUserID(ctx context.Context, userID string, status string, page, limit int) ([]entities.Order, int64, error) {
	var orders []entities.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Order{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if status != "" {
				return db.Where("status = ?", status)
			}
			return db
		}).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepositoryImpl) FindByPharmacyID(ctx context.Context, pharmacyID string, status string, page, limit int) ([]entities.Order, int64, error) {
	var orders []entities.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Order{}).Where("pharmacy_id = ?", pharmacyID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("pharmacy_id = ?", pharmacyID).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if status != "" {
				return db.Where("status = ?", status)
			}
			return db
		}).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepositoryImpl) FindAll(ctx context.Context, userID, pharmacyID, status, dateFrom, dateTo string, page, limit int) ([]entities.Order, int64, error) {
	var orders []entities.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Order{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if pharmacyID != "" {
		query = query.Where("pharmacy_id = ?", pharmacyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}
	query.Count(&total)

	offset := (page - 1) * limit
	dataQuery := r.db.WithContext(ctx).Preload("Items")
	if userID != "" {
		dataQuery = dataQuery.Where("user_id = ?", userID)
	}
	if pharmacyID != "" {
		dataQuery = dataQuery.Where("pharmacy_id = ?", pharmacyID)
	}
	if status != "" {
		dataQuery = dataQuery.Where("status = ?", status)
	}
	if dateFrom != "" {
		dataQuery = dataQuery.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		dataQuery = dataQuery.Where("created_at <= ?", dateTo)
	}
	err := dataQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepositoryImpl) Update(ctx context.Context, order *entities.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *OrderRepositoryImpl) GetNextOrderNumber(ctx context.Context) (int64, error) {
	var nextVal int64
	err := r.db.WithContext(ctx).Raw("SELECT nextval('orders.order_number_seq')").Scan(&nextVal).Error
	if err != nil {
		return 0, fmt.Errorf("error getting next order number: %w", err)
	}
	return nextVal, nil
}

func (r *OrderRepositoryImpl) GetOrderStats(ctx context.Context, dateFrom, dateTo string) (*repositories.OrderStats, error) {
	stats := &repositories.OrderStats{
		ByStatus: make(map[string]int64),
	}

	// Total orders and revenue
	baseQuery := r.db.WithContext(ctx).Model(&entities.Order{})
	if dateFrom != "" {
		baseQuery = baseQuery.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		baseQuery = baseQuery.Where("created_at <= ?", dateTo)
	}

	baseQuery.Count(&stats.TotalOrders)

	var totalRevenue float64
	r.db.WithContext(ctx).Model(&entities.Order{}).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if dateFrom != "" {
				db = db.Where("created_at >= ?", dateFrom)
			}
			if dateTo != "" {
				db = db.Where("created_at <= ?", dateTo)
			}
			return db
		}).
		Where("status NOT IN ?", []string{"cancelled", "pending_payment"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&totalRevenue)
	stats.TotalRevenue = totalRevenue

	if stats.TotalOrders > 0 {
		stats.AvgOrderValue = totalRevenue / float64(stats.TotalOrders)
	}

	// By status
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).Model(&entities.Order{}).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if dateFrom != "" {
				db = db.Where("created_at >= ?", dateFrom)
			}
			if dateTo != "" {
				db = db.Where("created_at <= ?", dateTo)
			}
			return db
		}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Top pharmacies
	type PharmacyStat struct {
		PharmacyID  string
		OrdersCount int64
		Revenue     float64
	}
	var pharmacyStats []PharmacyStat
	r.db.WithContext(ctx).Model(&entities.Order{}).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if dateFrom != "" {
				db = db.Where("created_at >= ?", dateFrom)
			}
			if dateTo != "" {
				db = db.Where("created_at <= ?", dateTo)
			}
			return db
		}).
		Where("status NOT IN ?", []string{"cancelled", "pending_payment"}).
		Select("pharmacy_id, COUNT(*) as orders_count, COALESCE(SUM(total), 0) as revenue").
		Group("pharmacy_id").
		Order("revenue DESC").
		Limit(10).
		Scan(&pharmacyStats)

	for _, ps := range pharmacyStats {
		stats.TopPharmacies = append(stats.TopPharmacies, repositories.TopPharmacyStat{
			PharmacyID:  ps.PharmacyID,
			OrdersCount: ps.OrdersCount,
			Revenue:     ps.Revenue,
		})
	}

	return stats, nil
}

var _ repositories.OrderRepository = (*OrderRepositoryImpl)(nil)
