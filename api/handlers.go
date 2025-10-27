package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"subscriptions-go/model"
	"subscriptions-go/service"
)

type Handler struct {
	svc *service.SubscriptionService
	log *logrus.Logger
}

func NewHandler(svc *service.SubscriptionService, log *logrus.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

type createReq struct {
	ServiceName string  `json:"service_name" binding:"required"`
	Price       int     `json:"price" binding:"gte=0"`
	UserID      string  `json:"user_id" binding:"required,uuid"`
	StartDate   string  `json:"start_date" binding:"required"` // MM-YYYY
	EndDate     *string `json:"end_date,omitempty"`            // MM-YYYY
}

// @Summary      Create a subscription
// @Description  Создает новую подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body  createReq  true  "Subscription info"
// @Success      201  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var r createReq
	if err := c.ShouldBindJSON(&r); err != nil {
		h.log.Warn("bad request create:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, _ := uuid.Parse(r.UserID)
	sd, err := parseMonthYear(r.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be in MM-YYYY format"})
		return
	}

	var ed *time.Time
	if r.EndDate != nil {
		t, err := parseMonthYear(*r.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be in MM-YYYY format"})
			return
		}
		ed = &t
	} // если r.EndDate == nil, ed останется nil — это вечная подписка

	sub := &model.Subscription{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      uid,
		StartDate:   sd,
		EndDate:     ed,
	}

	if err := h.svc.Create(sub); err != nil {
		h.log.Error("create error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// @Summary      Get a subscription by ID
// @Description  Получить подписку по ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /subscriptions/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sub, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// @Summary      List subscriptions
// @Description  Список всех подписок
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id       query   string  false "Filter by user ID"
// @Param        service_name  query   string  false "Filter by service name"
// @Success      200  {array}   model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	var uid *uuid.UUID
	if u := c.Query("user_id"); u != "" {
		parsed, err := uuid.Parse(u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be uuid"})
			return
		}
		uid = &parsed
	}

	var svcName *string
	if s := c.Query("service_name"); s != "" {
		svcName = &s
	}

	subs, err := h.svc.List(uid, svcName)
	if err != nil {
		h.log.Error("list error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// @Summary      Get subscription summary
// @Description  Суммарная информация по подпискам за период
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        start         query   string  true  "Start month MM-YYYY"
// @Param        end           query   string  true  "End month MM-YYYY"
// @Param        user_id       query   string  false "Filter by user ID"
// @Param        service_name  query   string  false "Filter by service name"
// @Success      200  {object}  map[string]int64
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/summary [get]
// @Summary      Get subscription summary
// @Description  Суммарная стоимость подписок за указанный период с учётом фильтров
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        start         query   string  true  "Start month MM-YYYY"
// @Param        end           query   string  true  "End month MM-YYYY"
// @Param        user_id       query   string  false "Filter by user ID"
// @Param        service_name  query   string  false "Filter by service name"
// @Success      200  {object}  map[string]int64
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/summary [get]
func (h *Handler) Summary(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required (MM-YYYY)"})
		return
	}

	periodStart, err := parseMonthYear(startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start must be MM-YYYY"})
		return
	}

	periodEnd, err := parseMonthYear(endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be MM-YYYY"})
		return
	}

	if periodEnd.Before(periodStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date cannot be before start date"})
		return
	}

	var userID *uuid.UUID
	if u := c.Query("user_id"); u != "" {
		uid, err := uuid.Parse(u)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be valid UUID"})
			return
		}
		userID = &uid
	}

	var serviceName *string
	if s := c.Query("service_name"); s != "" {
		serviceName = &s
	}

	// Получаем все подходящие подписки
	subs, err := h.svc.List(userID, serviceName)
	if err != nil {
		h.log.Error("summary error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var total int64
	for _, sub := range subs {
		subStart := sub.StartDate
		subEnd := sub.EndDate

		// Если подписка "вечная", считаем её до конца периода
		if subEnd == nil || subEnd.After(periodEnd) {
			subEnd = &periodEnd
		}

		// Если подписка начинается после конца периода — пропускаем
		if subStart.After(periodEnd) {
			continue
		}
		// Если подписка закончилась до начала периода — тоже пропускаем
		if subEnd.Before(periodStart) {
			continue
		}

		// Считаем пересечение дат
		overlapStart := maxTime(subStart, periodStart)
		overlapEnd := minTime(*subEnd, periodEnd)

		// Подсчёт месяцев пересечения
		months := monthsBetween(overlapStart, overlapEnd)
		if months > 0 {
			total += int64(months) * int64(sub.Price)
		}
	}

	c.JSON(http.StatusOK, gin.H{"total_rub": total})
}

// --- Вспомогательные функции ---
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func monthsBetween(start, end time.Time) int {
	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())
	return yearDiff*12 + monthDiff
}

func parseMonthYear(mmYYYY string) (time.Time, error) {
	t, err := time.Parse("01-2006", mmYYYY)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// @Summary      Update a subscription
// @Description  Обновляет подписку по ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id            path      string         true  "Subscription ID"
// @Param        subscription  body      createReq      true  "Subscription info"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var r createReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sd, err := parseMonthYear(r.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be MM-YYYY"})
		return
	}

	var ed *time.Time
	if r.EndDate != nil {
		t, err := parseMonthYear(*r.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be MM-YYYY"})
			return
		}
		ed = &t
	}

	sub, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	sub.ServiceName = r.ServiceName
	sub.Price = r.Price
	sub.UserID, _ = uuid.Parse(r.UserID)
	sub.StartDate = sd
	sub.EndDate = ed

	if err := h.svc.Update(sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// @Summary      Delete a subscription
// @Description  Удаляет подписку по ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.Status(http.StatusNoContent)
}
