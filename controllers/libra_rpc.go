package controllers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io.librablock.go/proto/admission_control"
	"io.librablock.go/proto/types"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"io.librablock.go/models"
)

const (
	MintProgramMd5   = "f0604842739be4f06a3d60227226858e"
	P2pProgramMd5    = "9b1b6bfc64fbe967a7f3d6606f7441d9"
	MintTransType    = "mint_transaction"
	P2pTransType     = "peer_to_peer_transaction"
	UnknownTransType = "unknown"
	DefaultAddress   = "ac.testnet.libra.org:8000"
)

type LibraRPC struct {
	Address string
}

func NewLibraRPC(address *string) LibraRPC {
	l := LibraRPC{}
	if address != nil {
		l.Address = *address
	} else {
		l.Address = DefaultAddress
	}

	return l
}

func BytesToHex(bytes []byte) string {
	dst := make([]byte, hex.EncodedLen(len(bytes)))
	n := hex.Encode(dst, bytes)
	return fmt.Sprintf("%s", dst[:n])
}

func HexToBytes(str string) ([]byte, error) {
	data, err := hex.DecodeString(str)
	return data, err
}

func HexToUint64(str string) (uint64, error) {
	bytes, err := HexToBytes(str)
	if err != nil {
		return 0, err
	}

	return uint64(binary.LittleEndian.Uint64(bytes)), nil
}

func (libra LibraRPC) GetLatestVersion() (uint64, error) {
	r, err := libra.updateToLatestLedgerRequest([]*types.RequestItem{libra.getTransactionsRequestMaker(0, 1, false)})

	if err != nil {
		return 0, err
	}

	return r.LedgerInfoWithSigs.LedgerInfo.Version, nil
}

func (libra LibraRPC) GetTransactions(version uint64, limit uint64, fetchEvents bool) (*[]models.BlockModel, error) {
	r, err := libra.updateToLatestLedgerRequest([]*types.RequestItem{libra.getTransactionsRequestMaker(version, limit, false)})

	if err != nil {
		fmt.Println(r)
		fmt.Println(err.Error())
		return nil, err
	}

	var res []models.BlockModel

	for _, x := range r.ResponseItems {
		switch val := x.ResponseItems.(type) {
		case *types.ResponseItem_GetTransactionsResponse:
			transactions := val.GetTransactionsResponse.TxnListWithProof.Transactions
			for idx, trans := range transactions {
				raw := types.RawTransaction{}
				err := proto.Unmarshal(trans.RawTxnBytes, &raw)
				if err != nil {
					return nil, err
				}

				result := models.BlockModel{}

				result.Version = version + uint64(idx)
				result.ExpirationAt = time.Unix(int64(raw.ExpirationTime), 0)
				result.Source = BytesToHex(raw.SenderAccount)
				result.GasPrice = raw.GasUnitPrice
				result.MaxGas = raw.MaxGasAmount
				result.SequenceNumber = raw.SequenceNumber
				result.PublicKey = BytesToHex(trans.SenderPublicKey)

				switch payload := raw.Payload.(type) {
				case *types.RawTransaction_Program:
					for _, arg := range payload.Program.Arguments {
						switch arg.Type {
						case types.TransactionArgument_U64:
							result.Amount = uint64(binary.LittleEndian.Uint64(arg.Data))
						case types.TransactionArgument_ADDRESS:
							result.Destination = BytesToHex(arg.Data)
						}
					}
					has := md5.Sum(payload.Program.Code)
					result.MD5 = fmt.Sprintf("%x", has)

					if result.MD5 == P2pProgramMd5 {
						result.Type = P2pTransType
					} else if result.MD5 == MintProgramMd5 {
						result.Type = MintTransType
					} else {
						result.Type = UnknownTransType
					}

				}

				res = append(res, result)
			}
		}
	}

	return &res, nil
}

func (libra LibraRPC) GetAccountState(address string) (*models.AccountModel, error) {
	result := models.AccountModel{}
	result.Address = address
	addressBytes, err := HexToBytes(address)

	if err != nil {
		return nil, err
	}

	r, err := libra.updateToLatestLedgerRequest([]*types.RequestItem{libra.getAccountStateRequestMaker(addressBytes)})

	if err != nil {
		return nil, err
	}

	for _, v := range r.ResponseItems {
		switch val := v.ResponseItems.(type) {
		case *types.ResponseItem_GetAccountStateResponse:
			blob := val.GetAccountStateResponse.AccountStateWithProof.Blob
			if blob.GetBlob() == nil {
				return nil, nil
			}

			str := BytesToHex(blob.Blob)
			magicStr := "100000001217da6c6b3e19f1825cfb2676daecce3bf3de03cf26647c78df00b371b25cc974500000020000000"
			idx := strings.Index(str, magicStr)
			it := len(magicStr) + idx
			addressLength := 64
			result.AuthenticationKey = str[it : it+addressLength]

			it += addressLength
			bitLength := 16
			var tmpArr []uint64
			for i := 0; i < 4; i++ {
				data, err := HexToUint64(str[it : it+bitLength])
				if err != nil {
					return nil, err
				}

				it += bitLength

				if i == 0 { //skip for a boolean param
					it += 2
				}

				tmpArr = append(tmpArr, data)
			}

			result.Balance = tmpArr[0]
			result.ReceivedEventCount = tmpArr[1]
			result.SentEventCount = tmpArr[2]
			result.SequenceNumber = tmpArr[3]

			break
		}
	}
	return &result, nil
}

func (libra LibraRPC) updateToLatestLedgerRequest(requests []*types.RequestItem) (*types.UpdateToLatestLedgerResponse, error) {
	conn, rpcErr := grpc.Dial(libra.Address, grpc.WithInsecure())
	if rpcErr != nil {
		log.Fatalf("did not connect: %v", rpcErr)
	}
	defer conn.Close()

	c := admission_control.NewAdmissionControlClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return c.UpdateToLatestLedger(
		ctx,
		&types.UpdateToLatestLedgerRequest{
			ClientKnownVersion: 0,
			RequestedItems:     requests,
		},
	)
}

func (libra LibraRPC) getTransactionsRequestMaker(version uint64, limit uint64, fetchEvents bool) *types.RequestItem {
	return &types.RequestItem{
		RequestedItems: &types.RequestItem_GetTransactionsRequest{
			GetTransactionsRequest: &types.GetTransactionsRequest{
				StartVersion: version,
				Limit:        limit,
				FetchEvents:  true,
			},
		}}
}

func (libra LibraRPC) getAccountStateRequestMaker(address []byte) *types.RequestItem {
	return &types.RequestItem{
		RequestedItems: &types.RequestItem_GetAccountStateRequest{
			GetAccountStateRequest: &types.GetAccountStateRequest{
				Address: address,
			},
		},
	}
}
