package delivery

import (
	"fmt"
)

type defaultSender struct {
	location string
	deliver  *defaultDeliverer
}

func newDefaultSender(deliver *defaultDeliverer, location string, options ...SenderOption) (*defaultSender, error) {
	sender := &defaultSender{
		location: location,
		deliver:  deliver,
	}
	for _, option := range options {
		if err := option.Apply(sender); err != nil {
			return nil, err
		}
	}

	return sender, nil
}

func (d *defaultSender) SendTo(i interface{}) error {
	deliver := d.deliver
	_, ok := deliver.receiverMap.Load(d.location)
	if !ok {
		return fmt.Errorf("no corresponding receiver: %s", d.location)
	}

	ch, ok := deliver.chanMap.Load(d.location)
	if !ok {
		return ErrLocationDeleted
	}

	switch ch := ch.(type) {
	case TransportChan:
		return d.sendTo(ch, i)
	default:
		return fmt.Errorf("unknow type")
	}
}

func (d *defaultSender) sendTo(senderChan SenderChan, data interface{}) error {
	select {
	case senderChan <- data:
		return nil
	default:
		return ErrReceiverBufferFulled
	}
}

func (d *defaultSender) SyncSendTo(i interface{}) (interface{}, error) {
	deliver := d.deliver
	receiver, ok := deliver.receiverMap.Load(d.location)
	if !ok {
		return nil, fmt.Errorf("no corresponding receiver: %s", d.location)
	}

	switch receiver := receiver.(type) {
	case *defaultReceiver:
		return receiver.syncProcess(i)
	default:
		return nil, fmt.Errorf("unknow type")
	}
}
