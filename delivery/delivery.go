package delivery

import (
	"fmt"
	"sync"
)

// Sender is an interface representing send message to Receiver.
type Sender interface {

	// SendTo asynchronously send data to Receiver based on 'location'.
	// An error occurs if the 'location' is deleted,
	// or if there is no corresponding Receiver,
	// or if the Receiver cannot handle so many messages
	SendTo(interface{}) error

	// SyncSendTo synchronize send data to Receiver based on 'location'.
	// Send data to Receiver until Receiver processing is complete.
	SyncSendTo(interface{}) (interface{}, error)
}

type ReceiveHandle func(interface{}) (interface{}, error)

type TransportChan = chan interface{}

type SenderChan = chan<- interface{}

type ReceiverChan = <-chan interface{}

// Receiver is an interface representing receive message from Sender.
type Receiver interface {
	// ToReceive accept data sent from Sender according to 'location'.
	// If there is no Sender or the 'location' is deleted,
	// an error will be reported.
	ToReceive() (interface{}, error)

	// ToSyncReceive Register a synchronous response function,
	// which will be called when the Sender sends data synchronously according to the 'location'.
	ToSyncReceive(ReceiveHandle)
}

type ReceiverOption interface {
	Apply(ReceiverConfig) error
}

// ReceiverConfig  receiver config
type ReceiverConfig interface {

	// SetBufferLength set receiver buffer.
	// This value depends on the processing speed of the receiver,
	// the faster the speed, the smaller the buffer can be set.
	// Throws an ErrReceiverBufferFulled if the sender is too fast
	SetBufferLength(len int)
}

// Deliverer is an interface representing manage sender and receiver.
// It can create Sender and Receiver based on 'location'.
// The Sender and Receiver 'locations' always need to be paired,
// and each 'location' needs to be different.
type Deliverer interface {
	// NewSender create Sender based on SenderOption and managed by Deliverer.
	NewSender(location string, options ...SenderOption) (Sender, error)

	// NewReceiver create Receiver based on ReceiverOption and managed by Deliverer.
	NewReceiver(location string, options ...ReceiverOption) (Receiver, error)

	// DeleteLocation delete 'location' from Deliverer.
	// It needs to ensure an idempotent interface.
	DeleteLocation(location string)
}

type SenderOption interface {
	Apply(SenderConfig) error
}

// SenderConfig  sender config
type SenderConfig interface {
}

type defaultDeliverer struct {
	senderMtx   sync.Mutex
	receiverMtx sync.Mutex
	senderMap   sync.Map
	receiverMap sync.Map
	chanMap     sync.Map
}

//New create a default Deliverer
func New() Deliverer {
	return &defaultDeliverer{}
}

func (d *defaultDeliverer) NewSender(location string, options ...SenderOption) (Sender, error) {
	// Make sure only one Sender can be created at the same time,
	// otherwise the same 'location' may be overwritten
	d.senderMtx.Lock()
	defer d.senderMtx.Unlock()

	_, ok := d.senderMap.Load(location)
	if ok {
		return nil, fmt.Errorf("sender locale already exists: %s", location)
	}

	sender, err := newDefaultSender(d, location, options...)
	if err != nil {
		return nil, err
	}
	d.senderMap.Store(location, sender)

	return sender, nil
}

func (d *defaultDeliverer) NewReceiver(location string, options ...ReceiverOption) (Receiver, error) {
	// Make sure only one Receiver can be created at the same time,
	// otherwise the same 'location' may be overwritten
	d.receiverMtx.Lock()
	defer d.receiverMtx.Unlock()

	_, ok := d.receiverMap.Load(location)
	if ok {
		return nil, fmt.Errorf("reciver location already exists: %s", location)
	}

	receiver, err := newDefaultReceiver(d, location, options...)
	if err != nil {
		return nil, err
	}

	d.receiverMap.Store(location, receiver)

	return receiver, nil
}

func (d *defaultDeliverer) DeleteLocation(location string) {
	d.receiverMtx.Lock()
	defer d.receiverMtx.Unlock()

	ch, ok := d.chanMap.LoadAndDelete(location)
	if !ok {
		return
	}

	switch ch := ch.(type) {
	case TransportChan:
		close(ch)
	}
}
