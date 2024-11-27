package handler

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/insomnius/wallet-event-loop/agregation"
	"github.com/insomnius/wallet-event-loop/entity"
	"github.com/insomnius/wallet-event-loop/repository"
	"github.com/labstack/echo/v4"
)

type TopUpRequest struct {
	Amount int `json:"amount"`
}

func TopUp(transactionAggregator *agregation.Transaction) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("current_user").(entity.UserToken).UserID

		var jsonBody TopUpRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, H{
				"errors": []H{
					{
						"detail": "bad json request",
					},
				},
			})

			return err
		}

		if err := transactionAggregator.TopUp(userID, jsonBody.Amount); err != nil {
			return c.JSON(http.StatusInternalServerError, H{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, H{"message": "TopUp successful"})
	}
}

func CheckBalance(walletRepo *repository.Wallet) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("current_user").(entity.UserToken).UserID

		// Get user balance
		userWallet, err := walletRepo.FindByUserID(userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, H{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, H{
			"user_id": userID,
			"balance": userWallet.Balance,
		})
	}
}

type TransferRequest struct {
	Amount int    `json:"amount"`
	To     string `json:"to"`
}

func Transfer(transactionAggregator *agregation.Transaction) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("current_user").(entity.UserToken).UserID

		var jsonBody TransferRequest

		if err := json.NewDecoder(c.Request().Body).Decode((&jsonBody)); err != nil {
			c.JSON(http.StatusBadRequest, H{
				"errors": []H{
					{
						"detail": "bad json request",
					},
				},
			})

			return err
		}

		if err := transactionAggregator.Transfer(userID, jsonBody.To, jsonBody.Amount); err != nil {
			if err == agregation.ErrInsuficientFound {
				return c.JSON(http.StatusBadRequest, H{"error": "Insufficient funds"})
			}
			return c.JSON(http.StatusInternalServerError, H{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, H{"message": "Transfer successful"})
	}
}

func TopTransfer(mutationRepo *repository.Mutation) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Getting the userID to filter transfers
		userID := c.Get("current_user").(entity.UserToken).UserID

		// Fetching the top 5 incoming and outgoing transactions
		mutations, err := mutationRepo.GetByUserID(userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		sort.Slice(mutations, func(i, j int) bool {
			return mutations[i].Amount > mutations[j].Amount
		})

		top5Incoming := make([]entity.Mutation, 0, 5)
		top5Outgoing := make([]entity.Mutation, 0, 5)

		for _, mu := range mutations {
			if mu.Type == entity.MutationTypeCredit && len(top5Incoming) < 5 {
				top5Incoming = append(top5Incoming, mu)
			} else if mu.Type == entity.MutationTypeDebit && len(top5Outgoing) < 5 {
				top5Outgoing = append(top5Outgoing, mu)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"incoming": top5Incoming,
			"outgoing": top5Outgoing,
		})
	}
}
