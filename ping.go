package ping4g

import (
	"errors"
	"net"
	"time"
)

type IcmpMessage struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	Body     IcmpMessageBodyerInterface
}
type IcmpMessageBodyerInterface interface {
	Len() int
	GetData() []byte
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

func (i *IcmpMessage) Marshal() ([]byte, error) {
	b := []byte{i.Type, i.Code, 0, 0}
	if i.Body != nil && i.Body.Len() > 0 {
		ib, err := i.Body.Marshal()
		if err != nil {
			return nil, err
		}
		b = append(b, ib...)
	}
	i.Checksum = i.GetCheckSum(b)
	// i.Checksum = CheckSum(b)
	b[2] = byte(i.Checksum & 0xff)
	b[3] = byte(i.Checksum >> 8)
	return b, nil
}

func (i *IcmpMessage) GetCheckSum(data []byte) uint16 {
	s := uint32(0)
	size := len(data)
	for i := 0; i < size-1; i += 2 {
		s += uint32(data[i+1])<<8 | uint32(data[i])
	}
	if size&1 == 1 {
		s += uint32(data[size-1])
	}
	s = s>>16 + s
	return uint16(^s)
}
func (i *IcmpMessage) Unmarshal(data []byte) error {
	msgLen := len(data)
	if msgLen < 4 {
		return errors.New("message too short")
	}
	i.Type = uint8(data[0])
	i.Code = uint8(data[1])
	i.Checksum = uint16(data[2])<<8 | uint16(data[3])
	if msgLen > 4 {
		i.Body = &IcmpMessageEcho{}
		i.Body.Unmarshal(data[4:])
	}

	return nil
}

type IcmpMessageEcho struct {
	ID       uint16
	Sequence uint16
	Data     []byte
}

func (e *IcmpMessageEcho) Len() int {
	if e == nil {
		return 0
	}
	return 4 + len(e.Data)
}
func (i *IcmpMessageEcho) GetData() []byte {
	return i.Data
}
func (e *IcmpMessageEcho) Marshal() ([]byte, error) {
	if e == nil {
		return nil, errors.New("IcmpMessageEcho is nil")
	}
	b := make([]byte, 4+len(e.Data))
	b[0], b[1] = byte(e.ID>>8), byte(e.ID&0xff)
	b[2], b[3] = byte(e.Sequence>>8), byte(e.Sequence&0xff)
	copy(b[4:], e.Data)
	return b, nil
}

func (i *IcmpMessageEcho) Unmarshal(data []byte) error {
	bodyLen := len(data)
	i.ID = uint16(data[0])<<8 | uint16(data[1])
	i.Sequence = uint16(data[2])<<8 | uint16(data[3])
	if bodyLen > 4 {
		i.Data = make([]byte, bodyLen-4)
		copy(i.Data, data[4:])
	}
	return nil
}
func Ping(addr string, timeout int) error {
	println(addr)
	c, err := net.Dial("ip4:icmp", addr)
	c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	defer c.Close()

	icmpEcho := &IcmpMessageEcho{
		ID:       1,
		Sequence: 1,
		Data:     []byte("Neu Spider"),
	}
	icmpMsg := &IcmpMessage{
		Type: 8,
		Code: 0,
		Body: icmpEcho,
	}

	data, _ := icmpMsg.Marshal()

	if _, err := c.Write(data); err != nil {
		panic(err)
	}
	recv := make([]byte, 20+len(data))
	var reply IcmpMessage
	for {
		if _, err := c.Read(recv); err != nil {
			panic(err)
			return err
		}
		if len(recv) > 20 {
			l := int(recv[0]&0x0f) << 2
			recv = recv[l:]
		}
		reply.Unmarshal(recv)
		println(string(reply.Body.GetData()))
		break
	}
	return err
}
