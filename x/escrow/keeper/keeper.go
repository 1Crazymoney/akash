package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ovrclk/akash/x/escrow/types"
)

type AccountHook func(sdk.Context, types.Account)
type PaymentHook func(sdk.Context, types.Payment)

type Keeper interface {
	AccountCreate(ctx sdk.Context, id types.AccountID, owner string, deposit sdk.Coin) error
	AccountDeposit(ctx sdk.Context, id types.AccountID, amount sdk.Coin) error
	AccountSettle(ctx sdk.Context, id types.AccountID) error
	AccountClose(ctx sdk.Context, id types.AccountID) error
	PaymentCreate(ctx sdk.Context, id types.AccountID, pid string, owner string, rate sdk.Coin) error
	PaymentWithdraw(ctx sdk.Context, id types.AccountID) error
	Paymentclose(ctx sdk.Context, id types.AccountID) error
	AddOnAccountClosedHook(AccountHook) Keeper
	AddOnPaymentClosedHook(PaymentHook) Keeper
}

func NewKeeper(cdc codec.BinaryMarshaler, skey sdk.StoreKey) Keeper {
	return &keeper{
		cdc:  cdc,
		skey: skey,
	}
}

type keeper struct {
	cdc  codec.BinaryMarshaler
	skey sdk.StoreKey

	hooks struct {
		onAccountClosed []AccountHook
		onPaymentClosed []PaymentHook
	}
}

func (k *keeper) AccountCreate(ctx sdk.Context, id types.AccountID, owner string, deposit sdk.Coin) error {
	return nil
}

func (k *keeper) AccountDeposit(ctx sdk.Context, id types.AccountID, amount sdk.Coin) error {
	return nil
}

func (k *keeper) AccountSettle(ctx sdk.Context, id types.AccountID) error {
	return nil
}

func (k *keeper) AccountClose(ctx sdk.Context, id types.AccountID) error {
	return nil
}

func (k *keeper) PaymentCreate(ctx sdk.Context, id types.AccountID, pid string, owner string, rate sdk.Coin) error {
	return nil
}

func (k *keeper) PaymentWithdraw(ctx sdk.Context, id types.AccountID) error {
	return nil
}

func (k *keeper) Paymentclose(ctx sdk.Context, id types.AccountID) error {
	return nil
}

func (k *keeper) AddOnAccountClosedHook(hook AccountHook) Keeper {
	k.hooks.onAccountClosed = append(k.hooks.onAccountClosed, hook)
	return k
}

func (k *keeper) AddOnPaymentClosedHook(hook PaymentHook) Keeper {
	k.hooks.onPaymentClosed = append(k.hooks.onPaymentClosed, hook)
	return k
}
