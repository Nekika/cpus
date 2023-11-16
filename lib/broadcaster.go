package lib

import (
	"errors"
	"sync"
)

type Broadcaster[T any] struct {
	subId  Id
	subsmx sync.Mutex
	subs   map[int]chan T
}

func NewBroadCaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subs: make(map[int]chan T),
	}
}

func (cub *Broadcaster[T]) Broadcast(ch <-chan T) error {
	for {
		val, ok := <-ch
		if !ok {
			return errors.New("channel closed")
		}

		cub.subsmx.Lock()

		for _, subch := range cub.subs {
			subch <- val
		}

		cub.subsmx.Unlock()
	}
}

func (cub *Broadcaster[T]) Register(ch chan T) (int, error) {
	cub.subId.Increment()

	cub.subsmx.Lock()
	defer cub.subsmx.Unlock()

	cub.subs[cub.subId.Value()] = ch

	return cub.subId.Value(), nil
}

func (cub *Broadcaster[T]) Revoke(id int) error {
	cub.subsmx.Lock()
	defer cub.subsmx.Unlock()

	delete(cub.subs, id)

	return nil
}
