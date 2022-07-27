package channel

import (
	"fmt"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v2/modules/core/04-channel/keeper"
	"github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
)

// InitGenesis initializes the ibc channel submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, channel := range gs.Channels {
		ch := types.NewChannel(channel.State, channel.Ordering, channel.Counterparty, channel.ConnectionHops, channel.Version)
		k.SetChannel(ctx, channel.PortId, channel.ChannelId, ch)
	}
	for _, ack := range gs.Acknowledgements {
		k.SetPacketAcknowledgement(ctx, ack.PortId, ack.ChannelId, ack.Sequence, ack.Data)
	}
	for _, commitment := range gs.Commitments {
		k.SetPacketCommitment(ctx, commitment.PortId, commitment.ChannelId, commitment.Sequence, commitment.Data)
	}
	for _, receipt := range gs.Receipts {
		k.SetPacketReceipt(ctx, receipt.PortId, receipt.ChannelId, receipt.Sequence)
	}
	for _, ss := range gs.SendSequences {
		k.SetNextSequenceSend(ctx, ss.PortId, ss.ChannelId, ss.Sequence)
	}
	for _, rs := range gs.RecvSequences {
		k.SetNextSequenceRecv(ctx, rs.PortId, rs.ChannelId, rs.Sequence)
	}
	for _, as := range gs.AckSequences {
		k.SetNextSequenceAck(ctx, as.PortId, as.ChannelId, as.Sequence)
	}
	k.SetNextChannelSequence(ctx, gs.NextChannelSequence)
}

// ExportGenesis returns the ibc channel submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	channels, err := k.GetAllChannels(ctx)
	if err != nil {
		panic(err)
	}

	acks, err := k.GetAllPacketAcks(ctx)
	if err != nil {
		panic(err)
	}

	commitments, err := k.GetAllPacketCommitments(ctx)
	if err != nil {
		panic(err)
	}

	receipts, err := k.GetAllPacketReceipts(ctx)
	if err != nil {
		panic(err)
	}

	sendSeqs, err := k.GetAllPacketSendSeqs(ctx)
	if err != nil {
		panic(err)
	}

	recvSeqs, err := k.GetAllPacketRecvSeqs(ctx)
	if err != nil {
		panic(err)
	}

	ackSeqs, err := k.GetAllPacketAckSeqs(ctx)
	if err != nil {
		panic(err)
	}

	return types.GenesisState{
		Channels:            channels,
		Acknowledgements:    acks,
		Commitments:         commitments,
		Receipts:            receipts,
		SendSequences:       sendSeqs,
		RecvSequences:       recvSeqs,
		AckSequences:        ackSeqs,
		NextChannelSequence: k.GetNextChannelSequence(ctx),
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
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", types.SubModuleName, err))
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
