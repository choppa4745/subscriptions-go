package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"subscriptions-go/model"
)

type SubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) *SubscriptionRepo { return &SubscriptionRepo{db: db} }

func (r *SubscriptionRepo) Create(sub *model.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *SubscriptionRepo) GetByID(id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	if err := r.db.First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriptionRepo) List(userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	db := r.db.Model(&model.Subscription{})

	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}
	if serviceName != nil {
		db = db.Where("service_name = ?", *serviceName)
	}

	var subs []*model.Subscription
	if err := db.Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func (r *SubscriptionRepo) Update(sub *model.Subscription) error {
	return r.db.Save(sub).Error
}

func (r *SubscriptionRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Subscription{}, "id = ?", id).Error
}

// Sum of price where subscription period intersects [periodStart, periodEnd].
// periodStart and periodEnd are dates representing first day of months.
func (r *SubscriptionRepo) SumPriceForPeriod(periodStart, periodEnd time.Time, userID *uuid.UUID, serviceName *string) (int64, error) {
	// SQL: sum price where (end_date is null and start_date <= periodEnd) OR (end_date is not null and start_date <= periodEnd and end_date >= periodStart)
	q := r.db.Model(&model.Subscription{}).Select("COALESCE(SUM(price),0) as total")

	q = q.Where(
		r.db.Where("end_date IS NULL AND start_date <= ?", periodEnd).
			Or("end_date IS NOT NULL AND start_date <= ? AND end_date >= ?", periodEnd, periodStart),
	)

	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if serviceName != nil {
		q = q.Where("service_name = ?", *serviceName)
	}

	var total int64
	if err := q.Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
