package common

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

func EncodeEvent(event *Event) ([]byte, error) {
	var buf bytes.Buffer

	// 使用小端字节序写入 Type（4字节）
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, uint32(event.Type))
	buf.Write(typeBytes)

	// 创建 gob 编码器
	enc := gob.NewEncoder(&buf)

	// 根据 Type 编码对应的 InnerEvent
	switch event.Type {
	case TradeEventType:
		trade := event.InnerEvent.(*TradeEvent)
		if err := enc.Encode(trade); err != nil {
			return nil, err
		}
	case TransferEventType:
		rawTransfer := event.InnerEvent.(*TransferEvent)
		if err := enc.Encode(rawTransfer); err != nil {
			return nil, err
		}
	case PoolEventType:
		pool := event.InnerEvent.(*PoolEvent)
		if err := enc.Encode(pool); err != nil {
			return nil, err
		}
	case TokenEventType:
		token := event.InnerEvent.(*TokenEvent)
		if err := enc.Encode(token); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown event type: %d", event.Type)
	}
	return buf.Bytes(), nil
}

func DecodeEvent(data []byte) (*Event, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	// 使用小端字节序读取 Type（4字节）
	eventType := EventType(binary.LittleEndian.Uint32(data[:4]))

	// 创建 gob 解码器，跳过前4个字节
	dec := gob.NewDecoder(bytes.NewReader(data[4:]))

	// 根据 Type 解码对应的 InnerEvent
	switch eventType {
	case TradeEventType:
		var innerTrade *TradeEvent
		if err := dec.Decode(&innerTrade); err != nil {
			return nil, fmt.Errorf("failed to decode trade event: %w", err)
		}
		return &Event{Type: eventType, InnerEvent: innerTrade}, nil
	case TransferEventType:
		var innerRawTransfer *TransferEvent
		if err := dec.Decode(&innerRawTransfer); err != nil {
			return nil, fmt.Errorf("failed to decode raw transfer event: %w", err)
		}
		return &Event{Type: eventType, InnerEvent: innerRawTransfer}, nil
	case PoolEventType:
		var innerPoolEvent *PoolEvent
		if err := dec.Decode(&innerPoolEvent); err != nil {
			return nil, fmt.Errorf("failed to decode pool event: %w", err)
		}
		return &Event{Type: eventType, InnerEvent: innerPoolEvent}, nil
	case TokenEventType:
		var innerTokenEvent *TokenEvent
		if err := dec.Decode(&innerTokenEvent); err != nil {
			return nil, fmt.Errorf("failed to decode token event: %w", err)
		}
		return &Event{Type: eventType, InnerEvent: innerTokenEvent}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %d", eventType)
	}
}

func init() {
	gob.Register(TradeEvent{})
	gob.Register(TransferEvent{})
	gob.Register(PoolEvent{})
	gob.Register(TokenEvent{})
}
