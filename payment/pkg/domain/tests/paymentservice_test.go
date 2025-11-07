package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"payment/pkg/domain/model"
	"payment/pkg/domain/service"
)

type testFixture struct {
	paymentService  service.Payment
	repo            *mockPaymentRepository
	eventDispatcher *mockEventDispatcher
}

func setup() testFixture {
	repo := &mockPaymentRepository{
		paymentStore: make(map[uuid.UUID]*model.Payment),
		walletStore:  make(map[uuid.UUID]*model.Wallet),
	}
	eventDispatcher := &mockEventDispatcher{}
	paymentService := service.NewPaymentService(repo, eventDispatcher)

	return testFixture{
		paymentService:  paymentService,
		repo:            repo,
		eventDispatcher: eventDispatcher,
	}
}

func TestPaymentService(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	orderID := uuid.Must(uuid.NewV7())
	paymentAmount := 99.99

	t.Run("Initiate payment", func(t *testing.T) {
		f := setup()
		_, _ = f.paymentService.CreateWallet(userID, 200.00)

		paymentID, err := f.paymentService.InitiatePayment(orderID, userID, paymentAmount)

		require.NoError(t, err)
		require.NotNil(t, f.repo.paymentStore[paymentID])
		require.Equal(t, model.Pending, f.repo.paymentStore[paymentID].Status)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.PaymentInitiated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Process successful payment", func(t *testing.T) {
		f := setup()
		initialBalance := 200.00
		_, _ = f.paymentService.CreateWallet(userID, initialBalance)
		paymentID, _ := f.paymentService.InitiatePayment(orderID, userID, paymentAmount)
		f.eventDispatcher.events = nil

		err := f.paymentService.ProcessPayment(paymentID)

		require.NoError(t, err)
		require.Equal(t, model.Completed, f.repo.paymentStore[paymentID].Status)
		userWallet, _ := f.repo.FindWalletByUserID(userID)
		require.Equal(t, initialBalance-paymentAmount, userWallet.Balance)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.PaymentCompleted{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Process failed payment due to insufficient funds", func(t *testing.T) {
		f := setup()
		initialBalance := 50.00
		_, _ = f.paymentService.CreateWallet(userID, initialBalance)
		paymentID, _ := f.paymentService.InitiatePayment(orderID, userID, paymentAmount)
		f.eventDispatcher.events = nil

		err := f.paymentService.ProcessPayment(paymentID)

		require.NoError(t, err)
		require.Equal(t, model.Failed, f.repo.paymentStore[paymentID].Status)
		require.NotNil(t, f.repo.paymentStore[paymentID].FailureReason)
		userWallet, _ := f.repo.FindWalletByUserID(userID)
		require.Equal(t, initialBalance, userWallet.Balance)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.PaymentFailed{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Fail to process already completed payment", func(t *testing.T) {
		f := setup()
		_, _ = f.paymentService.CreateWallet(userID, 200.00)
		paymentID, _ := f.paymentService.InitiatePayment(orderID, userID, paymentAmount)
		_ = f.paymentService.ProcessPayment(paymentID)
		f.eventDispatcher.events = nil

		err := f.paymentService.ProcessPayment(paymentID)

		require.ErrorIs(t, err, service.ErrPaymentAlreadyProcessed)
		require.Empty(t, f.eventDispatcher.events)
	})
}

var _ model.PaymentRepository = &mockPaymentRepository{}

type mockPaymentRepository struct {
	paymentStore map[uuid.UUID]*model.Payment
	walletStore  map[uuid.UUID]*model.Wallet
}

func (m *mockPaymentRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockPaymentRepository) StorePayment(payment *model.Payment) error {
	m.paymentStore[payment.ID] = payment
	return nil
}

func (m *mockPaymentRepository) FindPayment(id uuid.UUID) (*model.Payment, error) {
	if payment, ok := m.paymentStore[id]; ok {
		return payment, nil
	}
	return nil, model.ErrPaymentNotFound
}

func (m *mockPaymentRepository) StoreWallet(wallet *model.Wallet) error {
	m.walletStore[wallet.ID] = wallet
	return nil
}

func (m *mockPaymentRepository) FindWalletByUserID(userID uuid.UUID) (*model.Wallet, error) {
	for _, wallet := range m.walletStore {
		if wallet.UserID == userID {
			return wallet, nil
		}
	}
	return nil, model.ErrWalletNotFound
}

var _ service.EventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(event service.Event) error {
	m.events = append(m.events, event)
	return nil
}
