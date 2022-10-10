// Package mpcrpc is a client of mpc server, doing the sign and accept tasks.
package mpcrpc

import (
	"fmt"
	"math/big"
	"time"

	"github.com/anyswap/mpc-client/internal/tools"
	"github.com/anyswap/mpc-client/log"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	mpcToAddress       = "0x00000000000000000000000000000000000000dc"
	mpcWalletServiceID = 30400
)

var (
	mpcSigner = types.NewEIP155Signer(big.NewInt(mpcWalletServiceID))
	mpcToAddr = common.HexToAddress(mpcToAddress)

	mpcAPIPrefix  = "smpc_" // default prefix
	mpcSignType   = "ECDSA"
	mpcSignGroup  string
	mpcThreshold  string
	mpcMode       string
	mpcRPCAddress string
	mpcKeyWrapper *keystore.Key
	mpcUser       common.Address

	mpcRPCTimeout  = 10                // default to 10 seconds
	mpcSignTimeout = 120 * time.Second // default to 120 seconds
)

// MPCConfig mpc related config
type MPCConfig struct {
	APIPrefix    string
	RPCAddress   string
	RPCTimeout   uint64
	KeystoreFile string `json:"-"`
	PasswordFile string `json:"-"`

	NeedKeyStore bool `json:"-"`
	IsDKG        bool `json:"-"`

	SignTimeout uint64
	SignType    string // eg. ECDSA
	SignGroup   string
	Threshold   string
	Mode        *uint64 // 0:managed 1:private
}

// Init init mpc
func Init(mpcConfig *MPCConfig, isSign bool) {
	initRPC(mpcConfig)
	if isSign || mpcConfig.SignGroup != "" {
		initSign(mpcConfig)
	}
}

// SignWithKey sign by mpc node user with private key
func SignWithKey(message []byte) ([]byte, error) {
	return crypto.Sign(message, mpcKeyWrapper.PrivateKey)
}

func initRPC(mpcConfig *MPCConfig) {
	if mpcConfig.APIPrefix != "" {
		mpcAPIPrefix = mpcConfig.APIPrefix
	}
	if mpcConfig.RPCTimeout > 0 {
		mpcRPCTimeout = int(mpcConfig.RPCTimeout)
	}

	mpcRPCAddress = mpcConfig.RPCAddress
	if mpcRPCAddress == "" {
		log.Fatal("init mpc rpc failes, must specify mpc rpc url")
	}

	if mpcConfig.NeedKeyStore || mpcConfig.KeystoreFile != "" {
		key, err := tools.LoadKeyStore(mpcConfig.KeystoreFile, mpcConfig.PasswordFile)
		if err != nil {
			log.Fatal("load mpc user keystore failed", "err", err)
		}
		mpcKeyWrapper = key
		mpcUser = key.Address
		log.Info("load mpc user keystore success", "mpcUser", mpcUser.String())
	}

	log.Info("init mpc rpc success", "apiPrefix", mpcAPIPrefix, "rpcAddress", mpcRPCAddress, "rpcTimeout", mpcRPCTimeout)
}

func initSign(mpcConfig *MPCConfig) {
	if mpcConfig.SignTimeout > 0 {
		mpcSignTimeout = time.Duration(mpcConfig.SignTimeout * uint64(time.Second))
	}
	if mpcConfig.SignType != "" {
		mpcSignType = mpcConfig.SignType
	}
	mpcSignGroup = mpcConfig.SignGroup
	mpcThreshold = mpcConfig.Threshold
	mpcMode = fmt.Sprintf("%d", *mpcConfig.Mode)

	if mpcSignGroup == "" || mpcThreshold == "" {
		log.Fatal("init mpc sign failed, must specify sign group and threshold")
	}

	log.Info("init mpc sign success", "signType", mpcSignType, "signGroup", mpcSignGroup, "threshold", mpcThreshold, "mode", mpcMode, "signTimeout", mpcSignTimeout.String())
}
