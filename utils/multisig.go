package utils

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func AddMultiSigTypes(signTypes []SignatureType) ([]SignatureType, error) {
	enrichedSignTypes := make([]SignatureType, 0, len(signTypes)+2)
	enriched := false

	for _, signType := range signTypes {
		enrichedSignTypes = append(enrichedSignTypes, signType)
		if signType.Name == "hyperliquidChain" {
			enriched = true
			enrichedSignTypes = append(enrichedSignTypes,
				SignatureType{Name: "payloadMultiSigUser", Type: "address"},
				SignatureType{Name: "outerSigner", Type: "address"},
			)
		}
	}

	if !enriched {
		return nil, ErrInvalidSignatureChain
	}

	return enrichedSignTypes, nil
}

func AddMultiSigFields(action map[string]interface{}, payloadMultiSigUser, outerSigner string) map[string]interface{} {
	result := make(map[string]interface{}, len(action)+2)
	for k, v := range action {
		result[k] = v
	}

	result["payloadMultiSigUser"] = strings.ToLower(payloadMultiSigUser)
	result["outerSigner"] = strings.ToLower(outerSigner)

	return result
}

func SignMultiSigUserSignedActionPayload(
	wallet Wallet,
	action map[string]interface{},
	isMainnet bool,
	signTypes []SignatureType,
	txType string,
	payloadMultiSigUser string,
	outerSigner string,
) (Signature, error) {
	if !common.IsHexAddress(payloadMultiSigUser) {
		return Signature{}, fmt.Errorf("%w: payloadMultiSigUser", ErrInvalidAddress)
	}
	if !common.IsHexAddress(outerSigner) {
		return Signature{}, fmt.Errorf("%w: outerSigner", ErrInvalidAddress)
	}

	envelope := AddMultiSigFields(action, payloadMultiSigUser, outerSigner)

	enrichedSignTypes, err := AddMultiSigTypes(signTypes)
	if err != nil {
		return Signature{}, fmt.Errorf("enriching signature types: %w", err)
	}

	return SignUserSignedAction(wallet, envelope, enrichedSignTypes, txType, isMainnet)
}

func SignMultiSigL1ActionPayload(
	wallet Wallet,
	action interface{},
	isMainnet bool,
	vaultAddress string,
	timestamp uint64,
	payloadMultiSigUser string,
	outerSigner string,
) (Signature, error) {
	if !common.IsHexAddress(payloadMultiSigUser) {
		return Signature{}, fmt.Errorf("%w: payloadMultiSigUser", ErrInvalidAddress)
	}
	if !common.IsHexAddress(outerSigner) {
		return Signature{}, fmt.Errorf("%w: outerSigner", ErrInvalidAddress)
	}
	if vaultAddress != "" && !common.IsHexAddress(vaultAddress) {
		return Signature{}, fmt.Errorf("%w: vaultAddress", ErrInvalidAddress)
	}

	envelope := []interface{}{
		strings.ToLower(payloadMultiSigUser),
		strings.ToLower(outerSigner),
		action,
	}

	return SignL1Action(wallet, envelope, vaultAddress, timestamp, isMainnet)
}

func SignMultiSigAction(
	wallet Wallet,
	action map[string]interface{},
	isMainnet bool,
	vaultAddress string,
	nonce uint64,
) (Signature, error) {
	if vaultAddress != "" && !common.IsHexAddress(vaultAddress) {
		return Signature{}, fmt.Errorf("%w: vaultAddress", ErrInvalidAddress)
	}

	actionWithoutTag := make(map[string]interface{}, len(action)-1)
	for k, v := range action {
		if k != "type" {
			actionWithoutTag[k] = v
		}
	}

	multiSigActionHash, err := ActionHash(actionWithoutTag, vaultAddress, nonce)
	if err != nil {
		return Signature{}, fmt.Errorf("computing action hash: %w", err)
	}

	envelope := map[string]interface{}{
		"multiSigActionHash": hexutil.Encode(multiSigActionHash),
		"nonce":              nonce,
	}

	return SignUserSignedAction(
		wallet,
		envelope,
		MultiSigEnvelopeSignTypes,
		"HyperliquidTransaction:SendMultiSig",
		isMainnet,
	)
}

func CreateMultiSigAction(
	innerAction map[string]interface{},
	wallet Wallet,
	isMainnet bool,
	vaultAddress string,
	nonce uint64,
) (map[string]interface{}, Signature, error) {
	sig, err := SignMultiSigAction(wallet, innerAction, isMainnet, vaultAddress, nonce)
	if err != nil {
		return nil, Signature{}, fmt.Errorf("signing multi-sig action: %w", err)
	}

	envelope := map[string]interface{}{
		"type":         "sendMultiSig",
		"wallet":       strings.ToLower(wallet.Address().Hex()),
		"vaultAddress": vaultAddress,
		"nonce":        nonce,
		"action":       innerAction,
		"signature": map[string]interface{}{
			"r": sig.R,
			"s": sig.S,
			"v": sig.V,
		},
	}

	return envelope, sig, nil
}
