package agregation

import (
	"errors"

	"github.com/google/uuid"
	"github.com/insomnius/wallet-event-loop/db"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/kodefluence/aurelia"
)

var ErrUserAlreadyExists = errors.New("user already exists")

const encryptionKey = "somethingsecret-dont-try-this-really-lol"

type Authorization struct {
	walletRepo *repository.Wallet
	userRepo   *repository.User
	db         *db.Instance
}

func NewAuthorization(
	walletRepo *repository.Wallet,
	userRepo *repository.User,
	db *db.Instance,
) *Authorization {
	return &Authorization{
		walletRepo: walletRepo,
		userRepo:   userRepo,
		db:         db,
	}
}

func (a *Authorization) Register(email, password string) error {
	return a.db.Transaction(func(t *db.Transaction) error {
		existingUser, err := a.userRepo.FindByEmail(email, t)
		if err != db.ErrNotFound || existingUser != nil {
			return ErrUserAlreadyExists
		}

		userID := uuid.New().String()
		err = a.userRepo.Put(&entity.User{
			ID:       userID,
			Password: aurelia.Hash(password, encryptionKey),
			Email:    email,
		}, t)
		if err != nil {
			return err
		}

		err = a.walletRepo.Put(&entity.Wallet{
			ID:      uuid.New().String(),
			UserID:  userID,
			Balance: 0,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
