package main

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"log"
	"strings"
	"time"

	pb "./proto"
)

const (
	MintProgramMd5   = "f391bd853105aa8c8df4df6b72f57b87"
	P2pProgramMd5    = "6ae054e289074cdf4f90d366ffde798a"
	MintTransType    = "mint_transaction"
	P2pTransType     = "peer_to_peer_transaction"
	UnknownTransType = "unknown"
	DefaultAddress   = "ac.testnet.libra.org:8000"
)

type LibraRPC struct {
	Address string
}

func NewLibraRPC(address *string) LibraRPC  {
	l := LibraRPC{}
	if address != nil {
		l.Address = *address
	} else {
		l.Address = DefaultAddress
	}

	return l
}

func bytesToHex(bytes []byte) string {
	dst := make([]byte, hex.EncodedLen(len(bytes)))
	n := hex.Encode(dst, bytes)
	return fmt.Sprintf("%s", dst[:n])
}

func hexToBytes(str string) ([]byte, error) {
	data, err := hex.DecodeString(str)
	return data, err
}

func hexToUint64(str string) (uint64, error){
	bytes, err := hexToBytes(str)
	if err != nil {
		return 0, err
	}

	return uint64(binary.LittleEndian.Uint64(bytes)), nil
}

func (libra LibraRPC) GetLatestVersion() (uint64, error){
	r, err := libra.updateToLatestLedgerRequest([]*pb.RequestItem{ libra.getTransactionsRequestMaker(0,1, false)})

	if err != nil {
		return 0, err
	}

	return r.LedgerInfoWithSigs.LedgerInfo.Version, nil
}

func (libra LibraRPC) GetTransactions(version uint64, limit uint64, fetchEvents bool) (*[]BlockModel, error) {
	r, err := libra.updateToLatestLedgerRequest([]*pb.RequestItem{ libra.getTransactionsRequestMaker(version,limit, false)})

	if err != nil {
		return nil, err
	}

	var res []BlockModel

	for _, x := range r.ResponseItems {
		switch val := x.ResponseItems.(type) {
		case *pb.ResponseItem_GetTransactionsResponse:
			transactions := val.GetTransactionsResponse.TxnListWithProof.Transactions
					for idx, trans := range transactions {
						raw := pb.RawTransaction{}
						err := proto.Unmarshal(trans.RawTxnBytes, &raw)
						if err != nil {
							return nil, err
						}

						result := BlockModel{}

						result.Version = version + uint64(idx)
						result.ExpirationAt = time.Unix(int64(raw.ExpirationTime), 0)
						result.Source = bytesToHex(raw.SenderAccount)
						result.GasPrice = raw.GasUnitPrice
						result.MaxGas = raw.MaxGasAmount
						result.SequenceNumber = raw.SequenceNumber
						result.PublicKey = bytesToHex(trans.SenderPublicKey)

						switch payload := raw.Payload.(type) {
						case *pb.RawTransaction_Program:
							for _, arg := range payload.Program.Arguments {
								switch arg.Type {
								case pb.TransactionArgument_U64:
									result.Amount = uint64(binary.LittleEndian.Uint64(arg.Data))
								case pb.TransactionArgument_ADDRESS:
									result.Destination = bytesToHex(arg.Data)
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

func (libra LibraRPC) GetAccountState(address string) (*AccountModel, error) {
	result := AccountModel{}
	result.Address = address
	addressBytes, err := hexToBytes(address)

	if err != nil {
		return nil, err
	}

	r, err := libra.updateToLatestLedgerRequest([]*pb.RequestItem{ libra.getAccountStateRequestMaker(addressBytes)})

	if err != nil {
		return nil, err
	}

	for _, v := range r.ResponseItems {
		switch val := v.ResponseItems.(type) {
		case *pb.ResponseItem_GetAccountStateResponse:

			blob := val.GetAccountStateResponse.AccountStateWithProof.Blob
			str := bytesToHex(blob.Blob)
			magicStr := "100000001217da6c6b3e19f1825cfb2676daecce3bf3de03cf26647c78df00b371b25cc974400000020000000"
			idx := strings.Index(str, magicStr)

			it := len(magicStr) + idx
			addressLength := 64
			result.AuthenticationKey = str[it:it+addressLength]

			it += addressLength
			bitLength := 16
			var tmpArr []uint64
			for i:=0; i<4; i++ {
				data, err := hexToUint64(str[it:it+bitLength])
				if err != nil {
					return nil, err
				}

				it += bitLength
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

func (libra LibraRPC) updateToLatestLedgerRequest(requests []*pb.RequestItem) (*pb.UpdateToLatestLedgerResponse, error) {
	conn, rpcErr := grpc.Dial(libra.Address, grpc.WithInsecure())
	if rpcErr != nil {
		log.Fatalf("did not connect: %v", rpcErr)
	}
	defer conn.Close()

	c := pb.NewAdmissionControlClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return c.UpdateToLatestLedger(
		ctx,
		&pb.UpdateToLatestLedgerRequest{
			ClientKnownVersion: 0,
			RequestedItems:     requests,
		},
	)
}

func (libra LibraRPC) getTransactionsRequestMaker(version uint64, limit uint64, fetchEvents bool) *pb.RequestItem {
	return &pb.RequestItem{
		RequestedItems: &pb.RequestItem_GetTransactionsRequest{
			GetTransactionsRequest: &pb.GetTransactionsRequest{
				StartVersion: version,
				Limit:        limit,
				FetchEvents:  true,
			},
		}}
}

func (libra LibraRPC) getAccountStateRequestMaker(address []byte) *pb.RequestItem {
	return &pb.RequestItem{
		RequestedItems: &pb.RequestItem_GetAccountStateRequest{
			GetAccountStateRequest: &pb.GetAccountStateRequest{
				Address:address,
			},
		},
	}
}
