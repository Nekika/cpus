package lib

import "sync"

type Id struct {
	mx  sync.Mutex
	val int
}

func (i *Id) Increment() {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.val += 1
}

func (i *Id) Value() int {
	return i.val
}
