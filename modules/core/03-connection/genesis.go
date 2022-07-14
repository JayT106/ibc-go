package connection

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"

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

func ExportGenesisTo(ctx sdk.Context, k keeper.Keeper, exportPath string) error {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return err
	}

	var fileIndex = 0
	fn := fmt.Sprintf("%s%d", types.SubModuleName, fileIndex)
	filePath := path.Join(exportPath, fn)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fs := 0
	// write the genesis param to the file
	params := k.GetParams(ctx)
	encodedParam, err := params.Marshal()
	if err != nil {
		return err
	}

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(encodedParam)))
	n, err := f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	// write the nextClientSequence to the file
	seq := k.GetNextConnectionSequence(ctx)
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, seq)
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	// write the connections to the file
	connections, err := k.GetAllConnections(ctx)
	if err != nil {
		return err
	}
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(connections)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, connection := range connections {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := connection.Marshal()
			if err != nil {
				return err
			}
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(len(bz)))
			n, err := f.Write(b)
			if err != nil {
				return err
			}
			fs += n

			n, err = f.Write(bz)
			if err != nil {
				return err
			}
			fs += n

			// we limited the file size to 100M
			if fs > 100000000 {
				if err := f.Close(); err != nil {
					return err
				}

				fileIndex++
				f, err = os.Create(filePath)
				if err != nil {
					return err
				}

				fs = 0
			}
		}
	}

	// write the client connection path to the file
	connectionPaths, err := k.GetAllClientConnectionPaths(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(connectionPaths)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, connectionPath := range connectionPaths {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := connectionPath.Marshal()
			if err != nil {
				return err
			}
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(len(bz)))
			n, err := f.Write(b)
			if err != nil {
				return err
			}
			fs += n

			n, err = f.Write(bz)
			if err != nil {
				return err
			}
			fs += n

			// we limited the file size to 100M
			if fs > 100000000 {
				if err := f.Close(); err != nil {
					return err
				}

				fileIndex++
				f, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					return err
				}

				fs = 0
			}
		}
	}

	return nil
}
