package utils

import (
	"fmt"
	"time"
)

func OrderTypeToWire(orderType OrderType) (OrderTypeWire, error) {
	var result OrderTypeWire

	if orderType.Limit != nil {
		result.Limit = orderType.Limit
		return result, nil
	}

	if orderType.Trigger != nil {
		wirePrice, err := FloatToWire(orderType.Trigger.TriggerPx)
		if err != nil {
			return OrderTypeWire{}, fmt.Errorf("converting trigger price: %w", err)
		}

		result.Trigger = &TriggerOrderTypeWire{
			IsMarket:  orderType.Trigger.IsMarket,
			TriggerPx: wirePrice,
			TPSL:      orderType.Trigger.TPSL,
		}
		return result, nil
	}

	return OrderTypeWire{}, ErrInvalidOrderType
}

func OrderRequestToOrderWire(order OrderRequest, asset int) (OrderWire, error) {
	if err := order.Validate(); err != nil {
		return OrderWire{}, fmt.Errorf("invalid order request: %w", err)
	}

	wireType, err := OrderTypeToWire(order.OrderType)
	if err != nil {
		return OrderWire{}, fmt.Errorf("converting order type: %w", err)
	}

	wirePrice, err := FloatToWire(order.LimitPrice)
	if err != nil {
		return OrderWire{}, fmt.Errorf("converting limit price: %w", err)
	}

	wireSize, err := FloatToWire(order.Size)
	if err != nil {
		return OrderWire{}, fmt.Errorf("converting size: %w", err)
	}

	orderWire := OrderWire{
		Asset:      asset,
		IsBuy:      order.IsBuy,
		Price:      wirePrice,
		Size:       wireSize,
		ReduceOnly: order.ReduceOnly,
		Type:       wireType,
	}

	if order.Cloid != nil {
		cloid := order.Cloid.ToRaw()
		orderWire.Cloid = &cloid
	}

	return orderWire, nil
}

func OrderWiresToOrderAction(orderWires []OrderWire, builder string) OrderAction {
	return OrderAction{
		Type:     "order",
		Orders:   orderWires,
		Grouping: GroupingNA,
		Builder:  builder,
	}
}

func GetTimestampMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func BatchOrdersToWire(orders []OrderRequest, assetMap map[string]int) ([]OrderWire, error) {
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders provided")
	}

	wireOrders := make([]OrderWire, 0, len(orders))
	for _, order := range orders {
		asset, ok := assetMap[order.Coin]
		if !ok {
			return nil, fmt.Errorf("unknown asset: %s", order.Coin)
		}

		wireOrder, err := OrderRequestToOrderWire(order, asset)
		if err != nil {
			return nil, fmt.Errorf("converting order for %s: %w", order.Coin, err)
		}

		wireOrders = append(wireOrders, wireOrder)
	}

	return wireOrders, nil
}

func CreateLimitOrderType(tif TIF) LimitOrderType {
	return LimitOrderType{TIF: tif}
}

func CreateTriggerOrderType(triggerPrice float64, isMarket bool, tpsl TPSL) TriggerOrderType {
	return TriggerOrderType{
		TriggerPx: triggerPrice,
		IsMarket:  isMarket,
		TPSL:      tpsl,
	}
}

func CreateLimitOrder(
	coin string,
	isBuy bool,
	size float64,
	price float64,
	tif TIF,
	reduceOnly bool,
	cloid *Cloid,
) OrderRequest {
	return OrderRequest{
		Coin:       coin,
		IsBuy:      isBuy,
		Size:       size,
		LimitPrice: price,
		OrderType: OrderType{
			Limit: &LimitOrderType{TIF: tif},
		},
		ReduceOnly: reduceOnly,
		Cloid:      cloid,
	}
}

func CreateTriggerOrder(
	coin string,
	isBuy bool,
	size float64,
	limitPrice float64,
	triggerPrice float64,
	isMarket bool,
	tpsl TPSL,
	reduceOnly bool,
	cloid *Cloid,
) OrderRequest {
	return OrderRequest{
		Coin:       coin,
		IsBuy:      isBuy,
		Size:       size,
		LimitPrice: limitPrice,
		OrderType: OrderType{
			Trigger: &TriggerOrderType{
				TriggerPx: triggerPrice,
				IsMarket:  isMarket,
				TPSL:      tpsl,
			},
		},
		ReduceOnly: reduceOnly,
		Cloid:      cloid,
	}
}

func CreateModifyRequest(orderID int64, newOrder OrderRequest) ModifyRequest {
	return ModifyRequest{
		OrderID: orderID,
		Order:   newOrder,
	}
}

func CreateModifyRequestByCloid(cloid Cloid, newOrder OrderRequest) ModifyRequest {
	cloidCopy := cloid
	return ModifyRequest{
		Cloid: &cloidCopy,
		Order: newOrder,
	}
}

func ModifyRequestToWire(req ModifyRequest, assetMap map[string]int) (ModifyWire, error) {
	if err := req.Validate(); err != nil {
		return ModifyWire{}, fmt.Errorf("invalid modify request: %w", err)
	}

	asset, ok := assetMap[req.Order.Coin]
	if !ok {
		return ModifyWire{}, fmt.Errorf("unknown asset: %s", req.Order.Coin)
	}

	wireOrder, err := OrderRequestToOrderWire(req.Order, asset)
	if err != nil {
		return ModifyWire{}, fmt.Errorf("converting order: %w", err)
	}

	if req.OrderID <= 0 {
		return ModifyWire{}, fmt.Errorf("modification by Cloid not implemented")
	}

	return ModifyWire{
		OrderID: int(req.OrderID),
		Order:   wireOrder,
	}, nil
}
