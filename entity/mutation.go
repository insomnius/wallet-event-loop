package entity

type Mutation struct {
	ID     string
	Type   int // 0 debit 1 credit
	Amount int // amount of money
}
