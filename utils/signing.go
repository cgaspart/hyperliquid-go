package utils

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

func (d EIP712Domain) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":              d.Name,
		"version":           d.Version,
		"chainId":           d.ChainID,
		"verifyingContract": d.VerifyingContract,
	}
}

func DefaultExchangeDomain() EIP712Domain {
	return EIP712Domain{
		Name:              "Exchange",
		Version:           "1",
		ChainID:           1337,
		VerifyingContract: "0x0000000000000000000000000000000000000000",
	}
}

func DefaultHyperliquidDomain() EIP712Domain {
	return EIP712Domain{
		Name:              "HyperliquidSignTransaction",
		Version:           "1",
		ChainID:           421614,
		VerifyingContract: "0x0000000000000000000000000000000000000000",
	}
}

// SignatureTypesToMap converts SignatureType slices to the map format expected by EIP-712
func SignatureTypesToMap(types []SignatureType) []map[string]string {
	result := make([]map[string]string, len(types))
	for i, t := range types {
		result[i] = map[string]string{
			"name": t.Name,
			"type": t.Type,
		}
	}
	return result
}

// createEIP712TypedData creates a properly formatted EIP-712 typed data structure
func createEIP712TypedData(
	domain EIP712Domain,
	primaryType string,
	message map[string]interface{},
	types map[string][]SignatureType,
) map[string]interface{} {
	typesMap := make(map[string]interface{}, len(types)+1)
	for typeName, typeFields := range types {
		typesMap[typeName] = SignatureTypesToMap(typeFields)
	}
	typesMap["EIP712Domain"] = SignatureTypesToMap(EIP712DomainFields)

	return map[string]interface{}{
		"domain":      domain.ToMap(),
		"primaryType": primaryType,
		"types":       typesMap,
		"message":     message,
	}
}

// SignL1Action signs an L1 action
func SignL1Action(wallet Wallet, action interface{}, vaultAddress string, nonce uint64, isMainnet bool) (Signature, error) {
	hash, err := ActionHash(action, vaultAddress, nonce)
	if err != nil {
		return Signature{}, fmt.Errorf("computing action hash: %w", err)
	}

	agentMessage := ConstructPhantomAgent(hash, isMainnet)

	agentType := []SignatureType{
		{Name: "source", Type: "string"},
		{Name: "connectionId", Type: "bytes32"},
	}

	types := map[string][]SignatureType{
		"Agent": agentType,
	}

	typedData := createEIP712TypedData(
		DefaultExchangeDomain(),
		"Agent",
		agentMessage,
		types,
	)

	encodedData, err := msgpack.Marshal(typedData)
	if err != nil {
		return Signature{}, fmt.Errorf("encoding typed data: %w", err)
	}

	return wallet.SignMessage(encodedData)
}

func SignUserSignedAction(
	wallet Wallet,
	action map[string]interface{},
	payloadTypes []SignatureType,
	primaryType string,
	isMainnet bool,
) (Signature, error) {
	actionCopy := make(map[string]interface{}, len(action)+2)
	for k, v := range action {
		actionCopy[k] = v
	}

	actionCopy["signatureChainId"] = "0x66eee"
	if isMainnet {
		actionCopy["hyperliquidChain"] = "Mainnet"
	} else {
		actionCopy["hyperliquidChain"] = "Testnet"
	}

	types := map[string][]SignatureType{
		primaryType: payloadTypes,
	}

	typedData := createEIP712TypedData(
		DefaultHyperliquidDomain(),
		primaryType,
		actionCopy,
		types,
	)

	encodedData, err := msgpack.Marshal(typedData)
	if err != nil {
		return Signature{}, fmt.Errorf("encoding typed data: %w", err)
	}

	return wallet.SignMessage(encodedData)
}

func SignOrderAction(wallet Wallet, orderAction OrderAction, vaultAddress string, nonce uint64, isMainnet bool) (Signature, error) {
	actionMap, err := msgpack.Marshal(orderAction)
	if err != nil {
		return Signature{}, fmt.Errorf("marshalling order action: %w", err)
	}

	var action map[string]interface{}
	if err := msgpack.Unmarshal(actionMap, &action); err != nil {
		return Signature{}, fmt.Errorf("unmarshalling to map: %w", err)
	}

	return SignL1Action(wallet, action, vaultAddress, nonce, isMainnet)
}

func SignBatchOrderAction(wallet Wallet, orders []OrderWire, grouping GroupingType, builder string, vaultAddress string, nonce uint64, isMainnet bool) (Signature, error) {
	orderAction := OrderAction{
		Type:     "order",
		Orders:   orders,
		Grouping: grouping,
	}

	if builder != "" {
		orderAction.Builder = builder
	}

	return SignOrderAction(wallet, orderAction, vaultAddress, nonce, isMainnet)
}
