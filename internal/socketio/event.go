package socketio

import (
	"fmt"

	"github.com/auriora/onemount/internal/socketio/protocol"
)

const defaultPacketTracePayloadLimit = 512

// EnginePacketTrace captures raw Engine.IO packet metadata for diagnostics.
type EnginePacketTrace struct {
	Direction string
	Type      string
	Payload   string
}

// EngineMessageTrace captures decoded Socket.IO message metadata for diagnostics.
type EngineMessageTrace struct {
	Type      string
	Namespace string
	Event     string
	Payloads  []interface{}
}

func newEnginePacketTrace(direction string, limit int, p *protocol.Packet) *EnginePacketTrace {
	if p == nil {
		return nil
	}
	if limit <= 0 {
		limit = defaultPacketTracePayloadLimit
	}
	return &EnginePacketTrace{
		Direction: direction,
		Type:      packetTypeName(p.Type),
		Payload:   truncatePayload(p.Payload(), limit),
	}
}

func newEngineMessageTrace(limit int, m *protocol.Message) *EngineMessageTrace {
	if m == nil {
		return nil
	}
	return &EngineMessageTrace{
		Type:      messageTypeName(m.Type),
		Namespace: m.Namespace,
		Event:     m.Event,
		Payloads:  clonePayloads(m.Payloads),
	}
}

func packetTypeName(t protocol.PacketType) string {
	switch t {
	case protocol.PacketTypeOpen:
		return "open"
	case protocol.PacketTypeClose:
		return "close"
	case protocol.PacketTypePing:
		return "ping"
	case protocol.PacketTypePong:
		return "pong"
	case protocol.PacketTypeMessage:
		return "message"
	case protocol.PacketTypeUpgrade:
		return "upgrade"
	case protocol.PacketTypeNoop:
		return "noop"
	default:
		return fmt.Sprintf("unknown:%d", t)
	}
}

func messageTypeName(t protocol.MessageType) string {
	switch t {
	case protocol.MessageTypeConnect:
		return "connect"
	case protocol.MessageTypeDisconnect:
		return "disconnect"
	case protocol.MessageTypeEvent:
		return "event"
	case protocol.MessageTypeAck:
		return "ack"
	case protocol.MessageTypeError:
		return "error"
	case protocol.MessageTypeBinaryEvent:
		return "binary_event"
	case protocol.MessageTypeBinaryAck:
		return "binary_ack"
	default:
		return fmt.Sprintf("unknown:%d", t)
	}
}

func truncatePayload(payload string, limit int) string {
	if limit <= 0 || len(payload) <= limit {
		return payload
	}
	if limit <= 3 {
		return payload[:limit]
	}
	return payload[:limit-3] + "..."
}

func clonePayloads(payloads []interface{}) []interface{} {
	if len(payloads) == 0 {
		return nil
	}
	cloned := make([]interface{}, len(payloads))
	copy(cloned, payloads)
	return cloned
}
