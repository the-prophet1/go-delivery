package delivery

import (
	"fmt"
	"testing"
	"time"
)

const testLocation = "test"

func TestNew(t *testing.T) {
	deliver := New()
	sender, _ := deliver.NewSender(testLocation)
	receiver, _ := deliver.NewReceiver(testLocation)

	go func() {
		for {
			err := sender.SendTo("test msg")
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * 1)
			res, err := receiver.ToReceive()
			fmt.Println(res, err)
		}
	}()

	select {}
}

func TestSync(t *testing.T) {
	deliver := New()
	sender, _ := deliver.NewSender(testLocation)
	receiver, _ := deliver.NewReceiver(testLocation)
	receiver.ToSyncReceive(func(i interface{}) (interface{}, error) {
		time.Sleep(time.Second)
		fmt.Println("sync receive", i)
		return "complete", nil
	})

	for {
		res, err := sender.SyncSendTo("test msg")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}

}
