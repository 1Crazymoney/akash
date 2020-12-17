# Escrow Accounts and Payments

* [Overview](#overview)
* [Account Settlement](#account-settlement)
* [Models](#models)
* [Hooks](#hooks)

Escrow accounts are a mechanism that allow for time-based
payments from one bank account to another without block-by-block
micropayments.  They also support holding funds for an account
until an arbitrary event occurrs.

Escrow accounts are necessary in akash for two primary reasons:

1. Leases in Akash are priced in blocks - every new block, a payment
from the tenant (deployment owner) to the provider (lease holder)
is due.  Performance and security considerations prohibit the
naive approach of transferring tokens on every block.
1. Bidding on an order should not be free (for various reasons,
including performance and security).  Akash requires a deposit for every bid.
The deposit is returned to the bidder when the bid is closed.

## Overview

Escrow [accounts] are created with an arbitrary ID, an owner, and a balance.  The balance
is immediately transferred from the owner bank account to the escrow module [account].  [Accounts] may
have their balance increased by being deposited to after creation.

[Payments] represent transfers from the escrow account to another bank account.  They are
(currently) block-based - some amount is meant to be transferred for every block.  The amount
owed to the payment from the escrow [account] is subtracted from the escrow [account] balance
through a settlement process.

[Payments] may be withdrawn from, which transfers any undisbursed balance from the 
module account to the payment owner's bank account.

When an [account] or a [payment] is closed, any remaining balance will be transferred to
their respective owner accounts.

If at any time the amount owed to [payments] is greater than the remaining balance of the [account],
the account and all payments are closed with state `OVERDRAWN`.

Many actions invoke the settlement process and may cause the account to become overdrawn.

## Account Settlement

Account settlement is the process of updating internal accounting of the balances of [payments] for an
[account] to the current height.

Many actions trigger the account settlement process - it ensures an up-to-date ledger when
acting on the [account].

Account settlement goes as follows:

1. Determine `blockRate` - the amount owed for every block.
1. Determine `heightDelta` - the number of blocks since last settlement.
1. Determine `numFullBlocks` - the number of blocks that can be paid for in full.
1. Transfer amount for `numFullBlocks` to [payments].
1. If `numFullBlocks` is less than `heightDelta` ([account] overdrawn), distribute remaining balance
among [payments], weighted by each [payment]'s `rate`, and set [account] and [payments] to state `OVERDRAWN`.


```

SettlePayments(d Deposit, height uint, charges []Payment)

  // calculate rate per block for all payments
  blockRate := 0
  for p := range payments
    blockRate += p.Rate

  // height since last settlement
  heightDelta    := height - d.SettledAt

  // total amount to be paid
  totalTransfer  := blockRate * heightDelta

  // remaining balance in account
  accountBalance := d.Balance - d.Transferred

  // number of full blocks that can be paid for
  numFullBlocks  := MIN(accountBalance DIV blockRate,heightDelta)

  // transfer payments for full blocks
  for p := range payments
    p.Balance += p.Rate * numFullBlocks
  
  // not overdrawn (fully paid)
  if numFullBlocks == heightDelta
    d.SettledAt    = height
    d.Transferred += totalTransfer
    return

  // overdrawn
  remainingAmount := accountBalance - (blockRate * numFullBlocks)

  // distribute, balance, weighted by transfer amount.
  for p := range payments
    amount    := (p.Price / blockRate) * remainingAmount
    p.Balance += amount
    p.State    = OVERDRAWN

  d.Transferred += remainingAmount
  d.SettledAt    = height
  d.State        = OVERDRAWN
```

## Models

### Account

|Field|Description|
|---|---|
|`ID`|Unique ID of account.|
|`Owner`|Bank account address of owner.|
|`State`|Account state.|
|`Balance`|Amount deposited from owner bank account.|
|`Transferred`|Amount disbursed from account via payments.|
|`SettledAt`|Last block that payments were settled.|

#### Account State

|Name|
|---|
|`OPEN`|
|`CLOSED`|
|`OVERDRAWN`|

### Payment

|Field|Description|
|---|---|
|`AccountID`|Escrow [`Account`] ID.|
|`ID`|Unique (over `AccountID`) ID of payment.|
|`Owner`|Bank account address of owner.|
|`State`|Payment state.|
|`Rate`|Tokens per block to transfer.|
|`Balance`|Balance currently reserved for owner.|
|`Withdrawn`|Amount already withdrawn by owner.|

#### Payment State

|Name|
|---|
|`OPEN`|
|`CLOSED`|
|`OVERDRAWN`|

## Methods

### `AccountCreate`

Create an escrow account.  Funds are deposited
from the owner bank account to the escrow module account.

#### Arguments

|Field|Description|
|---|---|
|`ID`|Unique ID of account.|
|`Owner`|Bank account address of owner.|
|`Deposit`|Amount deposited from owner bank account.|

### `AccountDeposit`

Add funds to an escrow account.  Funds are transferred
from the owner bank account to the escrow module account.

#### Arguments

|Field|Description|
|---|---|
|`ID`|Unique ID of account.|
|`Balance`|Amount deposited from owner bank account.|

### `AccountSettle`

Re-calculate remaining account and payment balances.

#### Arguments

|Field|Description|
|---|---|
|`ID`|Unique ID of account.|

### `AccountClose`

Close account - settle and close payments, return remaining
account balance to owner bank account.

### `PaymentCreate`

Create a new payment.  The account will first
be settled; this method will fail if the account cannot be settled.

#### Arguments

|Field|Description|
|---|---|
|`AccountID`|Escrow [`Account`] ID.|
|`ID`|Unique (over `AccountID`) ID of payment.|
|`Owner`|Bank account address of owner.|
|`Rate`|Tokens per block to transfer.|

#### Invariants

* Account is in state `OPEN` after being settled.
* `ID` is unique.
* `Owner` exists.
* `Rate` is non-zero and account has funds for one block.

### `PaymentWithdraw`

Withdraw funds from a payment balance.  Account will
first be settled.

#### Arguments

|Field|Description|
|---|---|
|`AccountID`|Escrow [`Account`] ID.|
|`ID`|Unique (over `AccountID`) ID of payment.|

### `PaymentClose`

Close a payment.  Account will first be settled.

#### Arguments

|Field|Description|
|---|---|
|`AccountID`|Escrow [`Account`] ID.|
|`ID`|Unique (over `AccountID`) ID of payment.|

## Hooks

Hooks are callbacks that are registered by users of the escrow module that are
to be called on specific events.

### `OnAccountClosed`

Whenever an account is closed `OnAccountClosed(Account)` will be called.

### `OnPaymentClosed`

Whenever a payment is closed, `OnAccountClosed(Account)` will be called.

[Account Settlement]: #account-settlement
[`account`]: #account
[`accounts`]: #account
[`payment`]: #payment
[`payments`]: #payment
[account]: #account
[accounts]: #account
[payment]: #payment
[payments]: #payment
