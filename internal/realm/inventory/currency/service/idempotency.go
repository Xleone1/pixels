package service

import (
	"crypto/sha256"
	"fmt"

	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
)

// mutation maps grant parameters into repository data.
func mutation(params GrantParams, ledger bool) currencyrepo.Mutation {
	return currencyrepo.Mutation{
		OperationKey: params.OperationKey,
		RequestHash:  operationHash(params.PlayerID, params.CurrencyType, params.Amount, false, params.ActorID, params.Reason),
		PlayerID:     params.PlayerID, CurrencyType: params.CurrencyType, Amount: params.Amount,
		Ledger: ledger, Reason: params.Reason, ActorKind: params.ActorKind, ActorID: params.ActorID,
	}
}

// setMutation maps set parameters into repository data.
func setMutation(params SetParams, ledger bool) currencyrepo.Mutation {
	return currencyrepo.Mutation{
		OperationKey: params.OperationKey,
		RequestHash:  operationHash(params.PlayerID, params.CurrencyType, params.Amount, true, params.ActorID, params.Reason),
		PlayerID:     params.PlayerID, CurrencyType: params.CurrencyType, Amount: params.Amount,
		Ledger: ledger, Reason: params.Reason, ActorKind: params.ActorKind, ActorID: params.ActorID,
	}
}

// operationHash binds one idempotency key to immutable mutation input.
func operationHash(playerID int64, currencyType int32, amount int64, absolute bool, actorID *int64, reason string) string {
	actor := int64(0)
	if actorID != nil {
		actor = *actorID
	}
	value := fmt.Sprintf("%d:%d:%d:%t:%d:%s", playerID, currencyType, amount, absolute, actor, reason)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}
