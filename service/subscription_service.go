package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"subscriptions-go/model"
	"subscriptions-go/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepo
}

func NewSubscriptionService(r *repository.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: r}
}

func (s *SubscriptionService) GetByID(id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionService) Create(sub *model.Subscription) error {
	if sub.Price < 0 {
		return errors.New("price cannot be negative")
	}

	// Приводим даты к первому дню месяца
	sub.StartDate = time.Date(sub.StartDate.Year(), sub.StartDate.Month(), 1, 0, 0, 0, 0, sub.StartDate.Location())
	if sub.EndDate != nil {
		ed := time.Date(sub.EndDate.Year(), sub.EndDate.Month(), 1, 0, 0, 0, 0, sub.EndDate.Location())
		sub.EndDate = &ed
	}

	// Проверка пересечения подписок с тем же ServiceName
	existing, err := s.repo.List(&sub.UserID, &sub.ServiceName)
	if err != nil {
		return err
	}

	for _, e := range existing {
		eEnd := e.EndDate
		if eEnd == nil {
			temp := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
			eEnd = &temp
		}
		newEnd := sub.EndDate
		if newEnd == nil {
			temp := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
			newEnd = &temp
		}

		if sub.StartDate.Before(*eEnd) && (*newEnd).After(e.StartDate) {
			return fmt.Errorf("subscription for service %s overlaps with existing subscription", sub.ServiceName)
		}
	}

	return s.repo.Create(sub)
}

func (s *SubscriptionService) Update(sub *model.Subscription) error {
	if sub.Price < 0 {
		return errors.New("price cannot be negative")
	}

	sub.StartDate = time.Date(sub.StartDate.Year(), sub.StartDate.Month(), 1, 0, 0, 0, 0, sub.StartDate.Location())
	if sub.EndDate != nil {
		ed := time.Date(sub.EndDate.Year(), sub.EndDate.Month(), 1, 0, 0, 0, 0, sub.EndDate.Location())
		sub.EndDate = &ed
	}

	existing, err := s.repo.List(&sub.UserID, &sub.ServiceName)
	if err != nil {
		return err
	}

	for _, e := range existing {
		if e.ID == sub.ID {
			continue // пропускаем саму себя
		}
		eEnd := e.EndDate
		if eEnd == nil {
			temp := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
			eEnd = &temp
		}
		newEnd := sub.EndDate
		if newEnd == nil {
			temp := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
			newEnd = &temp
		}

		if sub.StartDate.Before(*eEnd) && (*newEnd).After(e.StartDate) {
			return fmt.Errorf("subscription for service %s overlaps with existing subscription", sub.ServiceName)
		}
	}

	return s.repo.Update(sub)
}

func (s *SubscriptionService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *SubscriptionService) List(userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	return s.repo.List(userID, serviceName)
}

// Sum с учётом вечных подписок и пересечений периодов
func (s *SubscriptionService) Sum(periodStart, periodEnd time.Time, userID *uuid.UUID, serviceName *string) (int64, error) {
	ps := time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	pe := time.Date(periodEnd.Year(), periodEnd.Month(), 1, 0, 0, 0, 0, time.UTC)

	subs, err := s.repo.List(userID, serviceName)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, sub := range subs {
		start := sub.StartDate
		end := sub.EndDate
		if end == nil || end.After(pe) {
			end = &pe
		}
		if start.After(pe) || end.Before(ps) {
			continue
		}

		overlapStart := maxTime(&start, &ps)
		overlapEnd := minTime(end, &pe)
		months := monthsBetween(overlapStart, overlapEnd)
		total += int64(months) * int64(sub.Price)
	}

	return total, nil
}

func maxTime(a, b *time.Time) *time.Time {
	if a.After(*b) {
		return a
	}
	return b
}

func minTime(a, b *time.Time) *time.Time {
	if a.Before(*b) {
		return a
	}
	return b
}

func monthsBetween(start, end *time.Time) int {
	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())
	return yearDiff*12 + monthDiff + 1
}
