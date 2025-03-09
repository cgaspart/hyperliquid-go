package utils

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrInvalidOrderType      = errors.New("invalid order type: neither limit nor trigger is specified")
	ErrPrecisionLoss         = errors.New("conversion would cause precision loss")
	ErrInvalidAddress        = errors.New("invalid ethereum address format")
	ErrInvalidSignatureChain = errors.New("hyperliquidChain missing from signature types")
)

const (
	PrecisionThreshold   = 1e-12
	DefaultDecimalPlaces = 8
	USDDecimalPlaces     = 6
)

type Cloid string

func (c Cloid) ToRaw() string {
	return string(c)
}

// ----- Order Type Definitions -----

// TIF (Time-in-Force) order type constants
type TIF string

const (
	TIFAlo TIF = "Alo" // At Limit Only
	TIFIoc TIF = "Ioc" // Immediate or Cancel
	TIFGtc TIF = "Gtc" // Good Till Canceled
)

type TPSL string

const (
	TPSLTakeProfit TPSL = "tp"
	TPSLStopLoss   TPSL = "sl"
)

type LimitOrderType struct {
	TIF TIF `json:"tif" msgpack:"tif"`
}

type TriggerOrderType struct {
	TriggerPx float64 `json:"triggerPx" msgpack:"triggerPx"`
	IsMarket  bool    `json:"isMarket" msgpack:"isMarket"`
	TPSL      TPSL    `json:"tpsl" msgpack:"tpsl"`
}

type TriggerOrderTypeWire struct {
	TriggerPx string `json:"triggerPx" msgpack:"triggerPx"`
	IsMarket  bool   `json:"isMarket" msgpack:"isMarket"`
	TPSL      TPSL   `json:"tpsl" msgpack:"tpsl"`
}

type OrderType struct {
	Limit   *LimitOrderType   `json:"limit,omitempty" msgpack:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty" msgpack:"trigger,omitempty"`
}

type OrderTypeWire struct {
	Limit   *LimitOrderType       `json:"limit,omitempty" msgpack:"limit,omitempty"`
	Trigger *TriggerOrderTypeWire `json:"trigger,omitempty" msgpack:"trigger,omitempty"`
}

type OrderRequest struct {
	Coin       string    `json:"coin" msgpack:"coin"`
	IsBuy      bool      `json:"is_buy" msgpack:"is_buy"`
	Size       float64   `json:"sz" msgpack:"sz"`
	LimitPrice float64   `json:"limit_px" msgpack:"limit_px"`
	OrderType  OrderType `json:"order_type" msgpack:"order_type"`
	ReduceOnly bool      `json:"reduce_only" msgpack:"reduce_only"`
	Cloid      *Cloid    `json:"cloid,omitempty" msgpack:"cloid,omitempty"`
}

func (o *OrderRequest) Validate() error {
	if o.Size <= 0 {
		return errors.New("size must be positive")
	}
	if o.LimitPrice <= 0 {
		return errors.New("limit price must be positive")
	}
	if o.OrderType.Limit == nil && o.OrderType.Trigger == nil {
		return ErrInvalidOrderType
	}
	if o.OrderType.Trigger != nil && o.OrderType.Trigger.TriggerPx <= 0 {
		return errors.New("trigger price must be positive")
	}
	return nil
}

type OrderWire struct {
	Asset      int           `json:"a" msgpack:"a"`           // Asset ID
	IsBuy      bool          `json:"b" msgpack:"b"`           // Buy/Sell flag
	Price      string        `json:"p" msgpack:"p"`           // Price as string
	Size       string        `json:"s" msgpack:"s"`           // Size as string
	ReduceOnly bool          `json:"r" msgpack:"r"`           // Reduce only flag
	Type       OrderTypeWire `json:"t" msgpack:"t"`           // Order type
	Cloid      *string       `json:"c,omitempty" msgpack:"c"` // Client order ID
}

// GroupingType represents different types of order grouping
type GroupingType string

const (
	GroupingNA           GroupingType = "na"
	GroupingNormalTPSL   GroupingType = "normalTpsl"
	GroupingPositionTPSL GroupingType = "positionTpsl"
)

// ModifyRequest represents a request to modify an order
type ModifyRequest struct {
	OrderID int64        `json:"oid,omitempty" msgpack:"oid,omitempty"`
	Cloid   *Cloid       `json:"cloid,omitempty" msgpack:"cloid,omitempty"`
	Order   OrderRequest `json:"order,omitempty" msgpack:"order,omitempty"`
}

func (m *ModifyRequest) Validate() error {
	if m.OrderID == 0 && m.Cloid == nil {
		return errors.New("either OrderID or Cloid must be specified")
	}
	return m.Order.Validate()
}

type ModifyWire struct {
	OrderID int       `json:"oid" msgpack:"oid"`
	Order   OrderWire `json:"order" msgpack:"order"`
}

// CancelRequest represents a request to cancel an order by ID
type CancelRequest struct {
	Coin    string `json:"coin" msgpack:"coin"`
	OrderID int64  `json:"oid" msgpack:"oid"`
}

func (c *CancelRequest) Validate() error {
	if c.Coin == "" {
		return errors.New("coin must be specified")
	}
	if c.OrderID <= 0 {
		return errors.New("order ID must be positive")
	}
	return nil
}

// CancelByCloidRequest represents a request to cancel an order by client ID
type CancelByCloidRequest struct {
	Coin  string `json:"coin" msgpack:"coin"`
	Cloid Cloid  `json:"cloid" msgpack:"cloid"`
}

func (c *CancelByCloidRequest) Validate() error {
	if c.Coin == "" {
		return errors.New("coin must be specified")
	}
	if c.Cloid == "" {
		return errors.New("cloid must be specified")
	}
	return nil
}

type OrderAction struct {
	Type     string       `json:"type" msgpack:"type"`
	Orders   []OrderWire  `json:"orders" msgpack:"orders"`
	Grouping GroupingType `json:"grouping" msgpack:"grouping"`
	Builder  string       `json:"builder,omitempty" msgpack:"builder,omitempty"`
}

// ----- Signature Definitions -----

// SignatureType represents EIP-712 field definitions for signatures
type SignatureType struct {
	Name string
	Type string
}

// Signature represents an ECDSA signature
type Signature struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint8  `json:"v"`
}

type Wallet interface {
	SignMessage(message []byte) (Signature, error)
	Address() common.Address
}

type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           int64
	VerifyingContract string
}

// Standard signature
var (
	USDSendSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	SpotTransferSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "token", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	WithdrawSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	USDClassTransferSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "toPerp", Type: "bool"},
		{Name: "nonce", Type: "uint64"},
	}

	ConvertToMultiSigUserSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "signers", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}

	MultiSigEnvelopeSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "multiSigActionHash", Type: "bytes32"},
		{Name: "nonce", Type: "uint64"},
	}

	AgentSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "agentAddress", Type: "address"},
		{Name: "agentName", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}

	BuilderFeeSignTypes = []SignatureType{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "maxFeeRate", Type: "string"},
		{Name: "builder", Type: "address"},
		{Name: "nonce", Type: "uint64"},
	}

	EIP712DomainFields = []SignatureType{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}
)
