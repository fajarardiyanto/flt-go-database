package interfaces

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type Encoder struct {
	Raw        io.Reader
	Encoding   Encoding
	exchange   string
	routingKey string
	context    context.Context
}

func NewEncoder(raw io.Reader, exchange, routingKey string, enc Encoding) Messages {
	return &Encoder{
		Raw:        raw,
		Encoding:   enc,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (c *Encoder) SetContext(ctx context.Context) {
	c.context = ctx
}

func (c *Encoder) Context() context.Context {
	return c.context
}

func (c *Encoder) Decode(data interface{}) error {
	ref := reflect.ValueOf(data).Elem()
	switch c.Encoding {
	case EncodingNone:
		bys := bytes.NewBuffer(nil)
		if _, err := io.Copy(bys, c.Raw); err == nil {
			switch ref.Kind() {
			case reflect.String:
				ref.SetString(bys.String())
			}
		}
		return nil
	case EncodingGob:
		return gob.NewDecoder(c.Raw).Decode(data)
	case EncodingBase64Gob:
		return gob.NewDecoder(c.Raw).Decode(data)
	case EncodingJSON:
		return json.NewDecoder(c.Raw).Decode(data)
	}
	return fmt.Errorf("encoding not supported")
}

func (c *Encoder) Encode(data interface{}) (res []byte, err error) {
	ref := reflect.TypeOf(data)
	switch c.Encoding {
	case EncodingNone:
		switch ref.Kind() {
		case reflect.Slice:
			if val, ok := data.([]byte); ok {
				return val, nil
			}
			return nil, fmt.Errorf("failed to covertion type data slice")
		case reflect.String:
			return []byte(data.(string)), nil
		}
	case EncodingGob:
		buffer := bytes.NewBuffer(nil)
		defer buffer.Reset()
		if err := gob.NewEncoder(buffer).Encode(data); err != nil {
			return res, err
		}
		return buffer.Bytes(), nil
	}

	return res, fmt.Errorf("encoding not supported")
}

func (c *Encoder) Exchange() string {
	return c.exchange
}

func (c *Encoder) RoutingKey() string {
	return c.routingKey
}
