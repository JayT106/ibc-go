package channel

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"

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
	// write the nextChqnnelSequence to the file
	seq := k.GetNextChannelSequence(ctx)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, seq)
	n, err := f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	// write the channels to the file
	channels, err := k.GetAllChannels(ctx)
	if err != nil {
		return err
	}
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(channels)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, channel := range channels {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := channel.Marshal()
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

	// write the channel packet acks to the file
	acks, err := k.GetAllPacketAcks(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(acks)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, ack := range acks {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := ack.Marshal()
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

	// write the channel packet commitments to the file
	commitments, err := k.GetAllPacketCommitments(ctx)
	if err != nil {
		return err
	}
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(commitments)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, commitment := range commitments {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := commitment.Marshal()
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

	// write the channel packet receipts to the file
	receipts, err := k.GetAllPacketReceipts(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(receipts)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, receipt := range receipts {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := receipt.Marshal()
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

	// write the channel packet send sequences to the file
	sendSeqs, err := k.GetAllPacketSendSeqs(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(sendSeqs)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, sendSeq := range sendSeqs {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := sendSeq.Marshal()
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

	// write the channel packet receive sequences to the file
	recvSeqs, err := k.GetAllPacketRecvSeqs(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(recvSeqs)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, recvSeq := range recvSeqs {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := recvSeq.Marshal()
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

	// write the channel packet ack sequences to the file
	ackSeqs, err := k.GetAllPacketAckSeqs(ctx)
	if err != nil {
		return err
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(ackSeqs)))
	n, err = f.Write(b)
	if err != nil {
		return err
	}
	fs += n

	for _, ackSeq := range ackSeqs {
		select {
		case <-ctx.Context().Done():
			return fmt.Errorf("context has been cancelled")
		default:
			bz, err := ackSeq.Marshal()
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
