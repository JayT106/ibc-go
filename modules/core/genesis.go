package ibc

import (
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/ibc-go/v2/modules/core/02-client"
	connection "github.com/cosmos/ibc-go/v2/modules/core/03-connection"
	channel "github.com/cosmos/ibc-go/v2/modules/core/04-channel"
	"github.com/cosmos/ibc-go/v2/modules/core/keeper"
	"github.com/cosmos/ibc-go/v2/modules/core/types"
)

// InitGenesis initializes the ibc state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, createLocalhost bool, gs *types.GenesisState) {
	client.InitGenesis(ctx, k.ClientKeeper, gs.ClientGenesis)
	connection.InitGenesis(ctx, k.ConnectionKeeper, gs.ConnectionGenesis)
	channel.InitGenesis(ctx, k.ChannelKeeper, gs.ChannelGenesis)
}

// ExportGenesis returns the ibc exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		ClientGenesis:     client.ExportGenesis(ctx, k.ClientKeeper),
		ConnectionGenesis: connection.ExportGenesis(ctx, k.ConnectionKeeper),
		ChannelGenesis:    channel.ExportGenesis(ctx, k.ChannelKeeper),
	}
}

func InitGenesisFrom(ctx sdk.Context, k keeper.Keeper, createLocalhost bool, importPath string) error {
	if err := client.InitGenesisFrom(ctx, k.ClientKeeper, importPath); err != nil {
		return err
	}

	if err := connection.InitGenesisFrom(ctx, k.ConnectionKeeper, importPath); err != nil {
		return err
	}

	if err := channel.InitGenesisFrom(ctx, k.ChannelKeeper, importPath); err != nil {
		return err
	}

	return nil
}

func ExportGenesisTo(ctx sdk.Context, k keeper.Keeper, exportPath string) error {
	if err := client.ExportGenesisTo(ctx, k.ClientKeeper, path.Join(exportPath, "client")); err != nil {
		return err
	}

	if err := connection.ExportGenesisTo(ctx, k.ConnectionKeeper, path.Join(exportPath, "connection")); err != nil {
		return err
	}

	if err := channel.ExportGenesisTo(ctx, k.ChannelKeeper, path.Join(exportPath, "channel")); err != nil {
		return err
	}

	return nil
}
