# go-socket-communication
Provides an easy way to communicate through sockets.


## Setting up a host
```
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	h := NewHost("/tmp/sock")
	go h.Start()

	ticker := time.NewTicker(time.Second)
	for {
		<- ticker.C
		msg := strconv.Itoa(rand.Int())
		fmt.Println(msg)
		h.SendMessage <- msg
	}
}
```

## Setting up a client
```
package main

import (
	"fmt"
	"time"
)

func main() {
	c := NewClient("/tmp/sock")
	c.Start()
	c.AddHandler <- PrintHandler{}
	defer c.Stop()

	for {
		time.Sleep(time.Minute)
	}
}

type PrintHandler struct {
}

func (p PrintHandler) Handle(msg string) {
	fmt.Println(msg)
}
```