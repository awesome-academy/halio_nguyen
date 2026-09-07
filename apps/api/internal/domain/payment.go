package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Payment status constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusRefunded   = "refunded"
)

// Payment method constants
const (
	PaymentMethodInternetBanking = "internet_banking"
	PaymentMethodBankTransfer    = "bank_transfer"
	PaymentMethodVNPay           = "vnpay"
	PaymentMethodMoMo            = "momo"
)

// Payment represents a monetary transaction for a booking.
type Payment struct {
	ID                     uuid.UUID        `json:"id"`
	BookingID              uuid.UUID        `json:"booking_id"`
	UserBankAccountID      *uuid.UUID       `json:"user_bank_account_id,omitempty"`
	UserBankAccount       *UserBankAccount `json:"user_bank_account,omitempty"`
	Amount                 float64          `json:"amount"`
	PaymentMethod          string           `json:"payment_method"`
	TransactionRef         *string          `json:"transaction_ref,omitempty"`
	BankName               *string          `json:"bank_name,omitempty"`
	Status                 string           `json:"status"`
	PaymentGatewayResponse json.RawMessage  `json:"payment_gateway_response,omitempty"`
	PaidAt                 *time.Time       `json:"paid_at,omitempty"`
	FailedReason           *string          `json:"failed_reason,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}
