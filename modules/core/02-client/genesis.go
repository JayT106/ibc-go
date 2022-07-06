package client

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v2/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v2/modules/core/exported"
)

// InitGenesis initializes the ibc client submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	k.SetParams(ctx, gs.Params)

	// Set all client metadata first. This will allow client keeper to overwrite client and consensus state keys
	// if clients accidentally write to ClientKeeper reserved keys.
	if len(gs.ClientsMetadata) != 0 {
		k.SetAllClientMetadata(ctx, gs.ClientsMetadata)
	}

	for _, client := range gs.Clients {
		cs, ok := client.ClientState.GetCachedValue().(exported.ClientState)
		if !ok {
			panic("invalid client state")
		}

		if !gs.Params.IsAllowedClient(cs.ClientType()) {
			panic(fmt.Sprintf("client state type %s is not registered on the allowlist", cs.ClientType()))
		}

		k.SetClientState(ctx, client.ClientId, cs)
	}

	for _, cs := range gs.ClientsConsensus {
		for _, consState := range cs.ConsensusStates {
			consensusState, ok := consState.ConsensusState.GetCachedValue().(exported.ConsensusState)
			if !ok {
				panic(fmt.Sprintf("invalid consensus state with client ID %s at height %s", cs.ClientId, consState.Height))
			}

			k.SetClientConsensusState(ctx, cs.ClientId, consState.Height, consensusState)
		}
	}

	k.SetNextClientSequence(ctx, gs.NextClientSequence)

	// NOTE: localhost creation is specifically disallowed for the time being.
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/7871
}

// ExportGenesis returns the ibc client submodule's exported genesis.
// NOTE: CreateLocalhost should always be false on export since a
// created localhost will be included in the exported clients.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	genClients, err := k.GetAllGenesisClients(ctx)
	if err != nil {
		panic(err)
	}
	clientsMetadata, err := k.GetAllClientMetadata(ctx, genClients)
	if err != nil {
		panic(err)
	}

	consensusStates, err := k.GetAllConsensusStates(ctx)
	if err != nil {
		panic(err)
	}

	return types.GenesisState{
		Clients:            genClients,
		ClientsMetadata:    clientsMetadata,
		ClientsConsensus:   consensusStates,
		Params:             k.GetParams(ctx),
		CreateLocalhost:    false,
		NextClientSequence: k.GetNextClientSequence(ctx),
	}
}

func ExportGenesisTo(ctx sdk.Context, k keeper.Keeper, exportPath string) error {
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return err
	}

	var fileIndex = 0
	fn := fmt.Sprintf("genesis%d", fileIndex)
	filePath := path.Join(exportPath, fn)
	f, err := os.Create(filePath)
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

	// write the createLocalhost to the file, the value is always false
	n, err = f.Write([]byte{0})
	if err != nil {
		return err
	}
	fs += n

	// write the nextClientSequence to the file
	seq := k.GetNextClientSequence(ctx)
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, seq)
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	// write the genesis clients to the file
	genClients, err := k.GetAllGenesisClients(ctx)
	if err != nil {
		return err
	}
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(genClients)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, genClient := range genClients {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := genClient.Marshal()
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

	// write the client metadata to the file
	clientsMetadata, err := k.GetAllClientMetadata(ctx, genClients)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(clientsMetadata)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, metadata := range clientsMetadata {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := metadata.Marshal()
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

	// write the client consensus state to the file
	consensusStates, err := k.GetAllConsensusStates(ctx)
	if err != nil {
		return err
	}
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(consensusStates)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, consensusState := range consensusStates {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := consensusState.Marshal()
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

	return nil
}
