package mpcrpc

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anyswap/mpc-client/log"
	"github.com/anyswap/mpc-client/mpcrpc/client"
)

// get mpc sign status error
var (
	ErrGetSignStatusTimeout = errors.New("getSignStatus timeout")
	ErrGetSignStatusFailed  = errors.New("getSignStatus failure")
	ErrGetDKGStatusTimeout  = errors.New("getDKGStatus timeout")
	ErrGetDKGStatusFailed   = errors.New("getDKGStatus failure")
)

const (
	successStatus = "Success"
)

func newWrongStatusError(subject, status, errInfo string) error {
	return fmt.Errorf("[%v] Wrong status \"%v\", err=\"%v\"", subject, status, errInfo)
}

func wrapPostError(method string, err error) error {
	return fmt.Errorf("[post] %v error, %w", mpcAPIPrefix+method, err)
}

func httpPost(result interface{}, method string, params ...interface{}) error {
	return client.RPCPostWithTimeout(mpcRPCTimeout, &result, mpcRPCAddress, mpcAPIPrefix+method, params...)
}

func httpPostTo(result interface{}, rpcAddress, method string, params ...interface{}) error {
	return client.RPCPostWithTimeout(mpcRPCTimeout, &result, rpcAddress, mpcAPIPrefix+method, params...)
}

// GetEnode call getEnode
func GetEnode(rpcAddr string) (string, error) {
	var result GetEnodeResp
	err := httpPostTo(&result, rpcAddr, "getEnode")
	if err != nil {
		return "", wrapPostError("getEnode", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("getEnode", result.Status, result.Error)
	}
	return result.Data.Enode, nil
}

// GetSignNonce call getSignNonce
func GetSignNonce(mpcUser, rpcAddr string) (uint64, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "getSignNonce", mpcUser)
	if err != nil {
		return 0, wrapPostError("getSignNonce", err)
	}
	if result.Status != successStatus {
		return 0, newWrongStatusError("getSignNonce", result.Status, result.Error)
	}
	bi, err := GetBigIntFromStr(result.Data.Result)
	if err != nil {
		return 0, fmt.Errorf("getSignNonce can't parse result as big int, %w", err)
	}
	return bi.Uint64(), nil
}

// GetSignStatus call getSignStatus
func GetSignStatus(key, rpcAddr string) (*SignStatus, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "getSignStatus", key)
	if err != nil {
		return nil, wrapPostError("getSignStatus", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getSignStatus", result.Status, "response error "+result.Error)
	}
	data := result.Data.Result
	var signStatus SignStatus
	err = json.Unmarshal([]byte(data), &signStatus)
	if err != nil {
		return nil, wrapPostError("getSignStatus", err)
	}
	switch signStatus.Status {
	case "Failure":
		log.Info("getSignStatus Failure", "keyID", key, "status", data)
		return nil, ErrGetSignStatusFailed
	case "Timeout":
		log.Info("getSignStatus Timeout", "keyID", key, "status", data)
		return nil, ErrGetSignStatusTimeout
	case successStatus:
		return &signStatus, nil
	default:
		return nil, newWrongStatusError("getSignStatus", signStatus.Status, "sign status error "+signStatus.Error)
	}
}

// GetAcceptList get accept list of 'user'
func GetAcceptList(user string) ([]*SignInfoData, error) {
	var result SignInfoResp
	err := httpPost(&result, "getCurNodeSignInfo", user)
	if err != nil {
		return nil, wrapPostError("getAcceptList", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getAcceptList", result.Status, result.Error)
	}
	return result.Data, nil
}

// GetCurNodeSignInfo call getCurNodeSignInfo
func GetCurNodeSignInfo() ([]*SignInfoData, error) {
	var result SignInfoResp
	err := httpPost(&result, "getCurNodeSignInfo", mpcKeyWrapper.Address.String())
	if err != nil {
		return nil, wrapPostError("getCurNodeSignInfo", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getCurNodeSignInfo", result.Status, result.Error)
	}
	return result.Data, nil
}

// Sign call sign
func Sign(raw, rpcAddr string) (string, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "sign", raw)
	if err != nil {
		return "", wrapPostError("sign", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("sign", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// AcceptSign call acceptSign
func AcceptSign(raw string) (string, error) {
	var result DataResultResp
	err := httpPost(&result, "acceptSign", raw)
	if err != nil {
		return "", wrapPostError("acceptSign", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("acceptSign", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// GetGroupByID call getGroupByID
func GetGroupByID(groupID, rpcAddr string) (*GroupInfo, error) {
	var result GetGroupByIDResp
	err := httpPostTo(&result, rpcAddr, "getGroupByID", groupID)
	if err != nil {
		return nil, wrapPostError("getGroupByID", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getGroupByID", result.Status, result.Error)
	}
	return result.Data, nil
}

// ReqDcrmAddr call reqDcrmAddr
func ReqDcrmAddr(raw, rpcAddr string) (string, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "reqDcrmAddr", raw)
	if err != nil {
		return "", wrapPostError("reqDcrmAddr", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("reqDcrmAddr", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// AcceptReqAddr call acceptReqAddr
func AcceptReqAddr(raw string) (string, error) {
	var result DataResultResp
	err := httpPost(&result, "acceptReqAddr", raw)
	if err != nil {
		return "", wrapPostError("acceptReqAddr", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("acceptReqAddr", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// GetReqAddrNonce call getReqAddrNonce
func GetReqAddrNonce(mpcUser, rpcAddr string) (uint64, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "getReqAddrNonce", mpcUser)
	if err != nil {
		return 0, wrapPostError("getReqAddrNonce", err)
	}
	if result.Status != successStatus {
		return 0, newWrongStatusError("getReqAddrNonce", result.Status, result.Error)
	}
	bi, err := GetBigIntFromStr(result.Data.Result)
	if err != nil {
		return 0, fmt.Errorf("getReqAddrNonce can't parse result as big int, %w", err)
	}
	return bi.Uint64(), nil
}

// GetReqAddrStatus call getReqAddrStatus
func GetReqAddrStatus(key, rpcAddr string) (*ReqAddrStatus, error) {
	var result DataResultResp
	err := httpPostTo(&result, rpcAddr, "getReqAddrStatus", key)
	if err != nil {
		return nil, wrapPostError("getReqAddrStatus", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getReqAddrStatus", result.Status, "response error "+result.Error)
	}
	data := result.Data.Result
	var reqAddrStatus ReqAddrStatus
	err = json.Unmarshal([]byte(data), &reqAddrStatus)
	if err != nil {
		return nil, wrapPostError("getReqAddrStatus", err)
	}
	switch reqAddrStatus.Status {
	case "Failure":
		log.Info("getReqAddrStatus Failure", "keyID", key, "status", data)
		return nil, ErrGetDKGStatusFailed
	case "Timeout":
		log.Info("getReqAddrStatus Timeout", "keyID", key, "status", data)
		return nil, ErrGetDKGStatusTimeout
	case successStatus:
		return &reqAddrStatus, nil
	default:
		return nil, newWrongStatusError("getReqAddrStatus", reqAddrStatus.Status, "sign status error "+reqAddrStatus.Error)
	}
}

// GetDKGAcceptList get dkg accept list
func GetDKGAcceptList(user string) ([]*ReqAddrInfoData, error) {
	var result ReqAddrInfoResp
	err := httpPost(&result, "getCurNodeReqAddrInfo", user)
	if err != nil {
		return nil, wrapPostError("getDKGAcceptList", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getDKGAcceptList", result.Status, result.Error)
	}
	return result.Data, nil
}

// GetCurNodeReqAddrInfo call getCurNodeReqAddrInfo
func GetCurNodeReqAddrInfo() ([]*ReqAddrInfoData, error) {
	var result ReqAddrInfoResp
	err := httpPost(&result, "getCurNodeReqAddrInfo", mpcKeyWrapper.Address.String())
	if err != nil {
		return nil, wrapPostError("getCurNodeReqAddrInfo", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getCurNodeReqAddrInfo", result.Status, result.Error)
	}
	return result.Data, nil
}

// GetAccounts call getAccounts
func GetAccounts(addr string) (*PubAccounts, error) {
	var result AccountInfoResp
	err := httpPost(&result, "getAccounts", addr, mpcMode)
	if err != nil {
		return nil, wrapPostError("getAccounts", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getAccounts", result.Status, "response error "+result.Error)
	}
	return result.Data.Result, nil
}
