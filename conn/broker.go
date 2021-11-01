package conn

import (
	"encoding/json"
	"log"
)

type handler func(msg map[string]interface{})

type Broker interface {
	AddHandler(code int, handl handler)
	Transfer(msg []byte)
}

type BrokerS struct {
	handlers map[int][]handler
}

func NewBroker() Broker {
	return &BrokerS{make(map[int][]handler)}
}

func (b *BrokerS) AddHandler(code int, handl handler) {
	if _, ok := b.handlers[code]; !ok {
		b.handlers[code] = make([]handler, 1)
	}
	b.handlers[code] = append(b.handlers[code], handl)
}

func (b *BrokerS) Transfer(msg []byte) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Println("Error while unmarshalling json message: ", err)
		return
	}

	c, ok := m["code"]
	if !ok {
		log.Println("Message does not contain code")
		return
	}
	code, ok := c.(int)
	if !ok {
		log.Println("Message code is not an integer")
		return
	}

	if handlers, ok := b.handlers[code]; ok {
		for _, h := range handlers {
			h(m)
		}
	}
}
