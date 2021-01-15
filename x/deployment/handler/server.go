package handler

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/ovrclk/akash/x/deployment/keeper"
	"github.com/ovrclk/akash/x/deployment/types"
	ekeeper "github.com/ovrclk/akash/x/escrow/keeper"
	etypes "github.com/ovrclk/akash/x/escrow/types"
)

var _ types.MsgServer = msgServer{}

const deploymentEscrowScope = "deployment"

type msgServer struct {
	deployment keeper.Keeper
	market     MarketKeeper
	escrow     ekeeper.Keeper
}

// NewServer returns an implementation of the deployment MsgServer interface
// for the provided Keeper.
func NewServer(k keeper.Keeper, mkeeper MarketKeeper) types.MsgServer {
	return &msgServer{deployment: k, market: mkeeper}
}

func (ms msgServer) CreateDeployment(goCtx context.Context, msg *types.MsgCreateDeployment) (*types.MsgCreateDeploymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.deployment.GetDeployment(ctx, msg.ID); found {
		return nil, types.ErrDeploymentExists
	}

	deployment := types.Deployment{
		DeploymentID: msg.ID,
		State:        types.DeploymentActive,
		Version:      msg.Version,
	}

	if err := types.ValidateDeploymentGroups(msg.Groups); err != nil {
		return nil, errors.Wrap(types.ErrInvalidGroups, err.Error())
	}

	groups := make([]types.Group, 0, len(msg.Groups))

	for idx, spec := range msg.Groups {
		groups = append(groups, types.Group{
			GroupID:   types.MakeGroupID(deployment.ID(), uint32(idx+1)),
			State:     types.GroupOpen,
			GroupSpec: spec,
		})
	}

	if err := ms.deployment.Create(ctx, deployment, groups); err != nil {
		return nil, errors.Wrap(types.ErrInternal, err.Error())
	}

	// create orders
	for _, group := range groups {
		if _, err := ms.market.CreateOrder(ctx, group.ID(), group.GroupSpec); err != nil {
			ctx.Logger().With("group", group.ID(), "error", err).Error("creating order")
			return &types.MsgCreateDeploymentResponse{}, err
		}
	}

	// todo: deposit

	if err := ms.escrow.AccountCreate(ctx, etypes.AccountID{
		Scope: deploymentEscrowScope,
		XID:   deployment.ID().String(),
	}, deployment.ID().Owner, sdk.NewCoin("XXX", sdk.NewInt(0))); err != nil {
		return &types.MsgCreateDeploymentResponse{}, err
	}

	return &types.MsgCreateDeploymentResponse{}, nil
}

func (ms msgServer) UpdateDeployment(goCtx context.Context, msg *types.MsgUpdateDeployment) (*types.MsgUpdateDeploymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	deployment, found := ms.deployment.GetDeployment(ctx, msg.ID)
	if !found {
		return nil, types.ErrDeploymentNotFound
	}

	if !bytes.Equal(msg.Version, deployment.Version) {
		deployment.Version = msg.Version
	}

	if err := ms.deployment.UpdateDeployment(ctx, deployment); err != nil {
		return nil, errors.Wrap(types.ErrInternal, err.Error())
	}

	return &types.MsgUpdateDeploymentResponse{}, nil
}

func (ms msgServer) CloseDeployment(goCtx context.Context, msg *types.MsgCloseDeployment) (*types.MsgCloseDeploymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	deployment, found := ms.deployment.GetDeployment(ctx, msg.ID)
	if !found {
		return nil, types.ErrDeploymentNotFound
	}

	if deployment.State == types.DeploymentClosed {
		return nil, types.ErrDeploymentClosed
	}

	if err := ms.escrow.AccountClose(ctx, etypes.AccountID{
		Scope: deploymentEscrowScope,
		XID:   deployment.ID().String(),
	}); err != nil {
		return &types.MsgCloseDeploymentResponse{}, err
	}

	// todo: maybe assert that it was closed.

	return &types.MsgCloseDeploymentResponse{}, nil
}

func (ms msgServer) CloseGroup(goCtx context.Context, msg *types.MsgCloseGroup) (*types.MsgCloseGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	group, found := ms.deployment.GetGroup(ctx, msg.ID)
	if !found {
		return nil, types.ErrGroupNotFound
	}

	// if Group already closed; return the validation error
	err := group.ValidateClosable()
	if err != nil {
		return nil, err
	}

	// Update the Group's state
	err = ms.deployment.OnCloseGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	ms.market.OnGroupClosed(ctx, group.ID())

	return &types.MsgCloseGroupResponse{}, nil
}
