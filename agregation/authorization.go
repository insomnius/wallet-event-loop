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
var ErrUserNotFound = errors.New("user not found")
var ErrAuthFailed = errors.New("authentication failed")

const encryptionKey = "somethingsecret-dont-try-this-really-lol"

type Authorization struct {
	walletRepo    *repository.Wallet
	userRepo      *repository.User
	userTokenRepo *repository.UserToken
	db            *db.Instance
}

func NewAuthorization(
	walletRepo *repository.Wallet,
	userRepo *repository.User,
	userTokenRepo *repository.UserToken,
	db *db.Instance,
) *Authorization {
	return &Authorization{
		walletRepo:    walletRepo,
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,
		db:            db,
	}
}

func (a *Authorization) Register(email, password string) error {
	return a.db.Transaction(func(t *db.Transaction) error {
		_, err := a.userRepo.FindByEmail(email, t)
		if err != db.ErrNotFound {
			return ErrUserAlreadyExists
		}

		userID := uuid.New().String()
		err = a.userRepo.Put(entity.User{
			ID:       userID,
			Password: aurelia.Hash(password, encryptionKey),
			Email:    email,
		}, t)
		if err != nil {
			return err
		}

		err = a.walletRepo.Put(entity.Wallet{
			ID:      uuid.New().String(),
			UserID:  userID,
			Balance: 0,
		}, t)
		if err != nil {
			return err
		}

		return nil
	})
}

func (a *Authorization) SignIn(email, password string) (string, error) {
	existingUser, err := a.userRepo.FindByEmail(email)
	if err != nil && err == db.ErrNotFound {
		return "", ErrUserNotFound
	}

	if !aurelia.Authenticate(password, encryptionKey, existingUser.Password) {
		return "", ErrAuthFailed
	}

	token := aurelia.Hash(uuid.New().String(), encryptionKey)
	if err := a.userTokenRepo.Put(entity.UserToken{
		UserID: existingUser.ID,
		Token:  token,
	}); err != nil {
		return "", err
	}
	return token, nil
}
