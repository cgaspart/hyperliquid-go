package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func SignUSDTransferAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["destination"]; !ok {
		return Signature{}, errors.New("missing required field: destination")
	}
	if _, ok := action["amount"]; !ok {
		return Signature{}, errors.New("missing required field: amount")
	}
	if _, ok := action["time"]; !ok {
		return Signature{}, errors.New("missing required field: time")
	}

	destination, ok := action["destination"].(string)
	if ok && !common.IsHexAddress(destination) {
		return Signature{}, fmt.Errorf("%w: destination", ErrInvalidAddress)
	}

	return SignUserSignedAction(wallet, action, USDSendSignTypes, "HyperliquidTransaction:UsdSend", isMainnet)
}

func SignSpotTransferAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["destination"]; !ok {
		return Signature{}, errors.New("missing required field: destination")
	}
	if _, ok := action["token"]; !ok {
		return Signature{}, errors.New("missing required field: token")
	}
	if _, ok := action["amount"]; !ok {
		return Signature{}, errors.New("missing required field: amount")
	}
	if _, ok := action["time"]; !ok {
		return Signature{}, errors.New("missing required field: time")
	}

	destination, ok := action["destination"].(string)
	if ok && !common.IsHexAddress(destination) {
		return Signature{}, fmt.Errorf("%w: destination", ErrInvalidAddress)
	}

	return SignUserSignedAction(wallet, action, SpotTransferSignTypes, "HyperliquidTransaction:SpotSend", isMainnet)
}

// SignWithdrawFromBridgeAction signs a withdraw from bridge action
func SignWithdrawFromBridgeAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	// Validate required fields
	if _, ok := action["destination"]; !ok {
		return Signature{}, errors.New("missing required field: destination")
	}
	if _, ok := action["amount"]; !ok {
		return Signature{}, errors.New("missing required field: amount")
	}
	if _, ok := action["time"]; !ok {
		return Signature{}, errors.New("missing required field: time")
	}

	// Validate destination address
	destination, ok := action["destination"].(string)
	if ok && !common.IsHexAddress(destination) {
		return Signature{}, fmt.Errorf("%w: destination", ErrInvalidAddress)
	}

	return SignUserSignedAction(wallet, action, WithdrawSignTypes, "HyperliquidTransaction:Withdraw", isMainnet)
}

func SignUSDClassTransferAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["amount"]; !ok {
		return Signature{}, errors.New("missing required field: amount")
	}
	if _, ok := action["toPerp"]; !ok {
		return Signature{}, errors.New("missing required field: toPerp")
	}
	if _, ok := action["nonce"]; !ok {
		return Signature{}, errors.New("missing required field: nonce")
	}

	return SignUserSignedAction(wallet, action, USDClassTransferSignTypes, "HyperliquidTransaction:UsdClassTransfer", isMainnet)
}

func SignConvertToMultiSigUserAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["signers"]; !ok {
		return Signature{}, errors.New("missing required field: signers")
	}
	if _, ok := action["nonce"]; !ok {
		return Signature{}, errors.New("missing required field: nonce")
	}

	return SignUserSignedAction(wallet, action, ConvertToMultiSigUserSignTypes, "HyperliquidTransaction:ConvertToMultiSigUser", isMainnet)
}

func SignAgentAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["agentAddress"]; !ok {
		return Signature{}, errors.New("missing required field: agentAddress")
	}
	if _, ok := action["agentName"]; !ok {
		return Signature{}, errors.New("missing required field: agentName")
	}
	if _, ok := action["nonce"]; !ok {
		return Signature{}, errors.New("missing required field: nonce")
	}

	// Validate agent address
	agentAddress, ok := action["agentAddress"].(string)
	if ok && !common.IsHexAddress(agentAddress) {
		return Signature{}, fmt.Errorf("%w: agentAddress", ErrInvalidAddress)
	}

	return SignUserSignedAction(wallet, action, AgentSignTypes, "HyperliquidTransaction:ApproveAgent", isMainnet)
}

func SignApproveBuilderFeeAction(wallet Wallet, action map[string]interface{}, isMainnet bool) (Signature, error) {
	if _, ok := action["maxFeeRate"]; !ok {
		return Signature{}, errors.New("missing required field: maxFeeRate")
	}
	if _, ok := action["builder"]; !ok {
		return Signature{}, errors.New("missing required field: builder")
	}
	if _, ok := action["nonce"]; !ok {
		return Signature{}, errors.New("missing required field: nonce")
	}

	builder, ok := action["builder"].(string)
	if ok && !common.IsHexAddress(builder) {
		return Signature{}, fmt.Errorf("%w: builder", ErrInvalidAddress)
	}

	return SignUserSignedAction(wallet, action, BuilderFeeSignTypes, "HyperliquidTransaction:ApproveBuilderFee", isMainnet)
}

func CreateUSDTransferAction(destination string, amount string, timestamp uint64) map[string]interface{} {
	return map[string]interface{}{
		"destination": strings.ToLower(destination),
		"amount":      amount,
		"time":        timestamp,
	}
}

func CreateSpotTransferAction(destination string, token string, amount string, timestamp uint64) map[string]interface{} {
	return map[string]interface{}{
		"destination": strings.ToLower(destination),
		"token":       token,
		"amount":      amount,
		"time":        timestamp,
	}
}

func CreateWithdrawAction(destination string, amount string, timestamp uint64) map[string]interface{} {
	return map[string]interface{}{
		"destination": strings.ToLower(destination),
		"amount":      amount,
		"time":        timestamp,
	}
}

func CreateUSDClassTransferAction(amount string, toPerp bool, nonce uint64) map[string]interface{} {
	return map[string]interface{}{
		"amount": amount,
		"toPerp": toPerp,
		"nonce":  nonce,
	}
}

func CreateAgentAction(agentAddress string, agentName string, nonce uint64) map[string]interface{} {
	return map[string]interface{}{
		"agentAddress": strings.ToLower(agentAddress),
		"agentName":    agentName,
		"nonce":        nonce,
	}
}

func CreateApproveBuilderFeeAction(maxFeeRate string, builder string, nonce uint64) map[string]interface{} {
	return map[string]interface{}{
		"maxFeeRate": maxFeeRate,
		"builder":    strings.ToLower(builder),
		"nonce":      nonce,
	}
}
