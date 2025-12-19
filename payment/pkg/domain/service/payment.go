package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"payment/pkg/domain/model"
)

var (
	ErrPaymentAlreadyProcessed = errors.New("payment has already been processed")
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Payment interface {
	CreateWallet(userID uuid.UUID, initialBalance float64) (uuid.UUID, error)
	InitiatePayment(orderID, userID uuid.UUID, amount float64) (uuid.UUID, error)
	ProcessPayment(paymentID uuid.UUID) error
}

func NewPaymentService(repo model.PaymentRepository, dispatcher EventDispatcher) Payment {
	return &paymentService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type paymentService struct {
	repo       model.PaymentRepository
	dispatcher EventDispatcher
}

func (s *paymentService) CreateWallet(userID uuid.UUID, initialBalance float64) (uuid.UUID, error) {
	walletID, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}
	currentTime := time.Now()
	wallet := &model.Wallet{
		ID:        walletID,
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
	return walletID, s.repo.StoreWallet(wallet)
}

func (s *paymentService) InitiatePayment(orderID, userID uuid.UUID, amount float64) (uuid.UUID, error) {
	if _, err := s.repo.FindWalletByUserID(userID); err != nil {
		return uuid.Nil, err
	}

	paymentID, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	payment := &model.Payment{
		ID:        paymentID,
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Status:    model.Pending,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := s.repo.StorePayment(payment); err != nil {
		return uuid.Nil, err
	}

	return paymentID, s.dispatcher.Dispatch(model.PaymentInitiated{
		PaymentID: paymentID,
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
	})
}

func (s *paymentService) ProcessPayment(paymentID uuid.UUID) error {
	payment, err := s.repo.FindPayment(paymentID)
	if err != nil {
		return err
	}

	if payment.Status != model.Pending {
		return ErrPaymentAlreadyProcessed
	}

	wallet, err := s.repo.FindWalletByUserID(payment.UserID)
	if err != nil {
		return err
	}

	if wallet.Balance < payment.Amount {
		reason := "insufficient funds"
		payment.Status = model.Failed
		payment.FailureReason = &reason
		payment.UpdatedAt = time.Now()
		if err := s.repo.StorePayment(payment); err != nil {
			return err
		}
		return s.dispatcher.Dispatch(model.PaymentFailed{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			FailureReason: reason,
		})
	}

	wallet.Balance -= payment.Amount
	wallet.UpdatedAt = time.Now()
	if err := s.repo.StoreWallet(wallet); err != nil {
		return err
	}

	payment.Status = model.Completed
	payment.UpdatedAt = time.Now()
	if err := s.repo.StorePayment(payment); err != nil {
		// TODO Откат списания? вернуть исходную ошибку, так как не получится сохранить событие
		return err
	}

	return s.dispatcher.Dispatch(model.PaymentCompleted{
		PaymentID: payment.ID,
		OrderID:   payment.OrderID,
		UserID:    payment.UserID,
	})
}
