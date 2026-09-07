package domain

import (
	"time"

	"github.com/google/uuid"
)

// DailyRevenueReport matches mv_daily_revenue_report materialized view.
type DailyRevenueReport struct {
	ReportDate        time.Time `json:"report_date"`
	TourID            uuid.UUID `json:"tour_id"`
	TourTitle         string    `json:"tour_title"`
	CategoryID        uuid.UUID `json:"category_id"`
	CategoryName      string    `json:"category_name"`
	TotalBookings     int64     `json:"total_bookings"`
	TotalParticipants int64     `json:"total_participants"`
	TotalRevenue      float64   `json:"total_revenue"`
}

// MonthlyRevenueReport matches mv_monthly_revenue_report materialized view.
type MonthlyRevenueReport struct {
	ReportMonth       string    `json:"report_month"`
	CategoryID        uuid.UUID `json:"category_id"`
	CategoryName      string    `json:"category_name"`
	TotalBookings     int64     `json:"total_bookings"`
	TotalParticipants int64     `json:"total_participants"`
	TotalRevenue      float64   `json:"total_revenue"`
}
