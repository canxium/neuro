package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/rpc/eth/shared"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

func (c beaconApiValidatorClient) proposeBeaconBlock(ctx context.Context, in *ethpb.GenericSignedBeaconBlock) (*ethpb.ProposeResponse, error) {
	var consensusVersion string
	var beaconBlockRoot [32]byte

	var err error
	var marshalledSignedBeaconBlockJson []byte
	blinded := false

	switch blockType := in.Block.(type) {
	case *ethpb.GenericSignedBeaconBlock_Phase0:
		consensusVersion = "phase0"
		beaconBlockRoot, err = blockType.Phase0.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for phase0 beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockPhase0(blockType.Phase0)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall phase0 beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_Altair:
		consensusVersion = "altair"
		beaconBlockRoot, err = blockType.Altair.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for altair beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockAltair(blockType.Altair)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall altair beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_Bellatrix:
		consensusVersion = "bellatrix"
		beaconBlockRoot, err = blockType.Bellatrix.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for bellatrix beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockBellatrix(blockType.Bellatrix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall bellatrix beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_BlindedBellatrix:
		blinded = true
		consensusVersion = "bellatrix"
		beaconBlockRoot, err = blockType.BlindedBellatrix.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for blinded bellatrix beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockBlindedBellatrix(blockType.BlindedBellatrix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall blinded bellatrix beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_Capella:
		consensusVersion = "capella"
		beaconBlockRoot, err = blockType.Capella.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for capella beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockCapella(blockType.Capella)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall capella beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_BlindedCapella:
		blinded = true
		consensusVersion = "capella"
		beaconBlockRoot, err = blockType.BlindedCapella.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for blinded capella beacon block")
		}

		marshalledSignedBeaconBlockJson, err = marshallBeaconBlockBlindedCapella(blockType.BlindedCapella)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshall blinded capella beacon block")
		}
	case *ethpb.GenericSignedBeaconBlock_Deneb:
		consensusVersion = "deneb"
		beaconBlockRoot, err = blockType.Deneb.Block.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "failed to compute block root for deneb beacon block")
		}
		// TODO: Fix this as part of beacon API PR
		signedBlock, err := shared.SignedBeaconBlockContentsDenebFromConsensus(&ethpb.SignedBeaconBlockAndBlobsDeneb{
			Block: blockType.Deneb,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert deneb beacon block contents")
		}
		marshalledSignedBeaconBlockJson, err = json.Marshal(signedBlock)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal deneb beacon block contents")
		}
	case *ethpb.GenericSignedBeaconBlock_BlindedDeneb:
		blinded = true
		consensusVersion = "deneb"
		//TODO: update for beaocn API
		//beaconBlockRoot, err = blockType.BlindedDeneb.SignedBlindedBlock.HashTreeRoot()
		//if err != nil {
		//	return nil, errors.Wrap(err, "failed to compute block root for blinded deneb beacon block")
		//}
		//signedBlock, err := shared.SignedBlindedBeaconBlockContentsDenebFromConsensus(blockType.BlindedDeneb)
		//if err != nil {
		//	return nil, errors.Wrap(err, "failed to convert blinded deneb beacon block contents")
		//}
		//marshalledSignedBeaconBlockJson, err = json.Marshal(signedBlock)
		//if err != nil {
		//	return nil, errors.Wrap(err, "failed to marshal blinded deneb beacon block contents")
		//}
	default:
		return nil, errors.Errorf("unsupported block type %T", in.Block)
	}

	var endpoint string

	if blinded {
		endpoint = "/eth/v1/beacon/blinded_blocks"
	} else {
		endpoint = "/eth/v1/beacon/blocks"
	}

	headers := map[string]string{"Eth-Consensus-Version": consensusVersion}
	if httpError, err := c.jsonRestHandler.Post(ctx, endpoint, headers, bytes.NewBuffer(marshalledSignedBeaconBlockJson), nil); err != nil {
		if httpError != nil && httpError.Code == http.StatusAccepted {
			// Error 202 means that the block was successfully broadcasted, but validation failed
			return nil, errors.Wrap(err, "block was successfully broadcasted but failed validation")
		}

		return nil, errors.Wrap(err, "failed to send POST data to REST endpoint")
	}

	return &ethpb.ProposeResponse{BlockRoot: beaconBlockRoot[:]}, nil
}

func marshallBeaconBlockPhase0(block *ethpb.SignedBeaconBlock) ([]byte, error) {
	signedBeaconBlockJson := &shared.SignedBeaconBlock{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BeaconBlock{
			Body: &shared.BeaconBlockBody{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
			},
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
		},
	}

	return json.Marshal(signedBeaconBlockJson)
}

func marshallBeaconBlockAltair(block *ethpb.SignedBeaconBlockAltair) ([]byte, error) {
	signedBeaconBlockAltairJson := &shared.SignedBeaconBlockAltair{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BeaconBlockAltair{
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
			Body: &shared.BeaconBlockBodyAltair{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
				SyncAggregate: &shared.SyncAggregate{
					SyncCommitteeBits:      hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeBits),
					SyncCommitteeSignature: hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeSignature),
				},
			},
		},
	}

	return json.Marshal(signedBeaconBlockAltairJson)
}

func marshallBeaconBlockBellatrix(block *ethpb.SignedBeaconBlockBellatrix) ([]byte, error) {
	signedBeaconBlockBellatrixJson := &shared.SignedBeaconBlockBellatrix{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BeaconBlockBellatrix{
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
			Body: &shared.BeaconBlockBodyBellatrix{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
				SyncAggregate: &shared.SyncAggregate{
					SyncCommitteeBits:      hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeBits),
					SyncCommitteeSignature: hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeSignature),
				},
				ExecutionPayload: &shared.ExecutionPayload{
					BaseFeePerGas: bytesutil.LittleEndianBytesToBigInt(block.Block.Body.ExecutionPayload.BaseFeePerGas).String(),
					BlockHash:     hexutil.Encode(block.Block.Body.ExecutionPayload.BlockHash),
					BlockNumber:   uint64ToString(block.Block.Body.ExecutionPayload.BlockNumber),
					ExtraData:     hexutil.Encode(block.Block.Body.ExecutionPayload.ExtraData),
					FeeRecipient:  hexutil.Encode(block.Block.Body.ExecutionPayload.FeeRecipient),
					GasLimit:      uint64ToString(block.Block.Body.ExecutionPayload.GasLimit),
					GasUsed:       uint64ToString(block.Block.Body.ExecutionPayload.GasUsed),
					LogsBloom:     hexutil.Encode(block.Block.Body.ExecutionPayload.LogsBloom),
					ParentHash:    hexutil.Encode(block.Block.Body.ExecutionPayload.ParentHash),
					PrevRandao:    hexutil.Encode(block.Block.Body.ExecutionPayload.PrevRandao),
					ReceiptsRoot:  hexutil.Encode(block.Block.Body.ExecutionPayload.ReceiptsRoot),
					StateRoot:     hexutil.Encode(block.Block.Body.ExecutionPayload.StateRoot),
					Timestamp:     uint64ToString(block.Block.Body.ExecutionPayload.Timestamp),
					Transactions:  jsonifyTransactions(block.Block.Body.ExecutionPayload.Transactions),
				},
			},
		},
	}

	return json.Marshal(signedBeaconBlockBellatrixJson)
}

func marshallBeaconBlockBlindedBellatrix(block *ethpb.SignedBlindedBeaconBlockBellatrix) ([]byte, error) {
	signedBeaconBlockBellatrixJson := &shared.SignedBlindedBeaconBlockBellatrix{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BlindedBeaconBlockBellatrix{
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
			Body: &shared.BlindedBeaconBlockBodyBellatrix{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
				SyncAggregate: &shared.SyncAggregate{
					SyncCommitteeBits:      hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeBits),
					SyncCommitteeSignature: hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeSignature),
				},
				ExecutionPayloadHeader: &shared.ExecutionPayloadHeader{
					BaseFeePerGas:    bytesutil.LittleEndianBytesToBigInt(block.Block.Body.ExecutionPayloadHeader.BaseFeePerGas).String(),
					BlockHash:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.BlockHash),
					BlockNumber:      uint64ToString(block.Block.Body.ExecutionPayloadHeader.BlockNumber),
					ExtraData:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ExtraData),
					FeeRecipient:     hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.FeeRecipient),
					GasLimit:         uint64ToString(block.Block.Body.ExecutionPayloadHeader.GasLimit),
					GasUsed:          uint64ToString(block.Block.Body.ExecutionPayloadHeader.GasUsed),
					LogsBloom:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.LogsBloom),
					ParentHash:       hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ParentHash),
					PrevRandao:       hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.PrevRandao),
					ReceiptsRoot:     hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ReceiptsRoot),
					StateRoot:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.StateRoot),
					Timestamp:        uint64ToString(block.Block.Body.ExecutionPayloadHeader.Timestamp),
					TransactionsRoot: hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.TransactionsRoot),
				},
			},
		},
	}

	return json.Marshal(signedBeaconBlockBellatrixJson)
}

func marshallBeaconBlockCapella(block *ethpb.SignedBeaconBlockCapella) ([]byte, error) {
	signedBeaconBlockCapellaJson := &shared.SignedBeaconBlockCapella{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BeaconBlockCapella{
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
			Body: &shared.BeaconBlockBodyCapella{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
				SyncAggregate: &shared.SyncAggregate{
					SyncCommitteeBits:      hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeBits),
					SyncCommitteeSignature: hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeSignature),
				},
				ExecutionPayload: &shared.ExecutionPayloadCapella{
					BaseFeePerGas: bytesutil.LittleEndianBytesToBigInt(block.Block.Body.ExecutionPayload.BaseFeePerGas).String(),
					BlockHash:     hexutil.Encode(block.Block.Body.ExecutionPayload.BlockHash),
					BlockNumber:   uint64ToString(block.Block.Body.ExecutionPayload.BlockNumber),
					ExtraData:     hexutil.Encode(block.Block.Body.ExecutionPayload.ExtraData),
					FeeRecipient:  hexutil.Encode(block.Block.Body.ExecutionPayload.FeeRecipient),
					GasLimit:      uint64ToString(block.Block.Body.ExecutionPayload.GasLimit),
					GasUsed:       uint64ToString(block.Block.Body.ExecutionPayload.GasUsed),
					LogsBloom:     hexutil.Encode(block.Block.Body.ExecutionPayload.LogsBloom),
					ParentHash:    hexutil.Encode(block.Block.Body.ExecutionPayload.ParentHash),
					PrevRandao:    hexutil.Encode(block.Block.Body.ExecutionPayload.PrevRandao),
					ReceiptsRoot:  hexutil.Encode(block.Block.Body.ExecutionPayload.ReceiptsRoot),
					StateRoot:     hexutil.Encode(block.Block.Body.ExecutionPayload.StateRoot),
					Timestamp:     uint64ToString(block.Block.Body.ExecutionPayload.Timestamp),
					Transactions:  jsonifyTransactions(block.Block.Body.ExecutionPayload.Transactions),
					Withdrawals:   jsonifyWithdrawals(block.Block.Body.ExecutionPayload.Withdrawals),
				},
				BlsToExecutionChanges: jsonifyBlsToExecutionChanges(block.Block.Body.BlsToExecutionChanges),
			},
		},
	}

	return json.Marshal(signedBeaconBlockCapellaJson)
}

func marshallBeaconBlockBlindedCapella(block *ethpb.SignedBlindedBeaconBlockCapella) ([]byte, error) {
	signedBeaconBlockCapellaJson := &shared.SignedBlindedBeaconBlockCapella{
		Signature: hexutil.Encode(block.Signature),
		Message: &shared.BlindedBeaconBlockCapella{
			ParentRoot:    hexutil.Encode(block.Block.ParentRoot),
			ProposerIndex: uint64ToString(block.Block.ProposerIndex),
			Slot:          uint64ToString(block.Block.Slot),
			StateRoot:     hexutil.Encode(block.Block.StateRoot),
			Body: &shared.BlindedBeaconBlockBodyCapella{
				Attestations:      jsonifyAttestations(block.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(block.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(block.Block.Body.Deposits),
				Eth1Data:          jsonifyEth1Data(block.Block.Body.Eth1Data),
				Graffiti:          hexutil.Encode(block.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(block.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(block.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(block.Block.Body.VoluntaryExits),
				SyncAggregate: &shared.SyncAggregate{
					SyncCommitteeBits:      hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeBits),
					SyncCommitteeSignature: hexutil.Encode(block.Block.Body.SyncAggregate.SyncCommitteeSignature),
				},
				ExecutionPayloadHeader: &shared.ExecutionPayloadHeaderCapella{
					BaseFeePerGas:    bytesutil.LittleEndianBytesToBigInt(block.Block.Body.ExecutionPayloadHeader.BaseFeePerGas).String(),
					BlockHash:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.BlockHash),
					BlockNumber:      uint64ToString(block.Block.Body.ExecutionPayloadHeader.BlockNumber),
					ExtraData:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ExtraData),
					FeeRecipient:     hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.FeeRecipient),
					GasLimit:         uint64ToString(block.Block.Body.ExecutionPayloadHeader.GasLimit),
					GasUsed:          uint64ToString(block.Block.Body.ExecutionPayloadHeader.GasUsed),
					LogsBloom:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.LogsBloom),
					ParentHash:       hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ParentHash),
					PrevRandao:       hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.PrevRandao),
					ReceiptsRoot:     hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.ReceiptsRoot),
					StateRoot:        hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.StateRoot),
					Timestamp:        uint64ToString(block.Block.Body.ExecutionPayloadHeader.Timestamp),
					TransactionsRoot: hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.TransactionsRoot),
					WithdrawalsRoot:  hexutil.Encode(block.Block.Body.ExecutionPayloadHeader.WithdrawalsRoot),
				},
				BlsToExecutionChanges: jsonifyBlsToExecutionChanges(block.Block.Body.BlsToExecutionChanges),
			},
		},
	}

	return json.Marshal(signedBeaconBlockCapellaJson)
}
