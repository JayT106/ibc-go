package connection

import (
	"fmt"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v2/modules/core/03-connection/keeper"
	"github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
)

// InitGenesis initializes the ibc connection submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, connection := range gs.Connections {
		conn := types.NewConnectionEnd(connection.State, connection.ClientId, connection.Counterparty, connection.Versions, connection.DelayPeriod)
		k.SetConnection(ctx, connection.Id, conn)
	}
	for _, connPaths := range gs.ClientConnectionPaths {
		k.SetClientConnectionPaths(ctx, connPaths.ClientId, connPaths.Paths)
	}
	k.SetNextConnectionSequence(ctx, gs.NextConnectionSequence)
	k.SetParams(ctx, gs.Params)
}

// ExportGenesis returns the ibc connection submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	connections, err := k.GetAllConnections(ctx)
	if err != nil {
		panic(err)
	}

	connectionPaths, err := k.GetAllClientConnectionPaths(ctx)
	if err != nil {
		panic(err)
	}

	return types.GenesisState{
		Connections:            connections,
		ClientConnectionPaths:  connectionPaths,
		NextConnectionSequence: k.GetNextConnectionSequence(ctx),
		Params:                 k.GetParams(ctx),
	}
}

func InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, k keeper.Keeper, importPath string) error {
	fp := path.Join(importPath, fmt.Sprintf("genesis_%s.bin", types.SubModuleName))
	f, err := os.OpenFile(fp, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	bz := make([]byte, fi.Size())
	if _, err := f.Read(bz); err != nil {
		return err
	}

	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return err
	}

	InitGenesis(ctx, k, gs)
	return nil
}

func ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, k keeper.Keeper, exportPath string) error {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return err
	}

	fp := path.Join(exportPath, fmt.Sprintf("genesis_%s.bin", types.SubModuleName))
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := ExportGenesis(ctx, k)
	bz := cdc.MustMarshalJSON(&gs)
	if _, err := f.Write(bz); err != nil {
		return err
	}

	return nil
}
