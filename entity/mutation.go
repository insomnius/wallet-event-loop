package entity

type MutationType int

const MutationTypeDebit MutationType = 0
const MutationTypeCredit MutationType = 1

type Mutation struct {
	ID       string
	WalletID string       // relation to wallet
	UserID   string       // denormalize mutation data with userID
	Type     MutationType // 0 credit 1 debit
	Amount   int          // amount of money
}
