package delivery

import "fmt"

const defaultReceiveBufferLength = 5

type defaultReceiver struct {
	location     string
	deliver      *defaultDeliverer
	handle       ReceiveHandle
	bufferLength int
}

type ReceiveOption struct {
	BufferLength int
}

func (option *ReceiveOption) Apply(r ReceiverConfig) error {
	r.SetBufferLength(option.BufferLength)
	return nil
}

func newDefaultReceiver(deliver *defaultDeliverer, location string, option ...ReceiverOption) (*defaultReceiver, error) {
	receiver := &defaultReceiver{
		location:     location,
		deliver:      deliver,
		bufferLength: defaultReceiveBufferLength,
	}

	for _, receiveOption := range option {
		if err := receiveOption.Apply(receiver); err != nil {
			return nil, err
		}
	}

	// set receiver's channel size and hand it over to
	//the Deliverer for management
	ch := make(TransportChan, receiver.bufferLength)
	receiver.deliver.chanMap.Store(location, ch)

	return receiver, nil
}

func (d *defaultReceiver) ToReceive() (interface{}, error) {
	deliver := d.deliver
	_, ok := deliver.senderMap.Load(d.location)
	if !ok {
		return nil, fmt.Errorf("no corresponding sender: %s", d.location)
	}

	ch, ok := deliver.chanMap.Load(d.location)
	if !ok {
		return nil, ErrLocationDeleted
	}

	switch ch := ch.(type) {
	case TransportChan:
		return d.toReceive(ch), nil
	default:
		return nil, fmt.Errorf("unknow type")
	}
}

func (d *defaultReceiver) toReceive(receiverChan ReceiverChan) interface{} {
	res := <-receiverChan
	return res
}

func (d *defaultReceiver) ToSyncReceive(handle ReceiveHandle) {
	d.handle = handle
}

func (d *defaultReceiver) syncProcess(i interface{}) (interface{}, error) {
	return d.handle(i)
}

func (d *defaultReceiver) SetBufferLength(len int) {
	d.bufferLength = len
}
