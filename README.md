# mullCtrl
**mullCtrl** is an easy-to-use library that allows controlling the Mullvad client via its cli interface.

The primary use case is for iterating over a list of servers easily to provide a number of private
connections without connecting to a server more than once.

## Demo
https://user-images.githubusercontent.com/12067516/144759458-1b7ae80b-a10e-4a38-b11e-ae5fea66e1ea.mp4

## Usage

**API**

```go
client.GetStatus()
client.GetAccount()
client.GetServers()

client.ConnectToServer(s Server)
client.IsConnected()

client.IterateAllRandom()
client.IterateCountryRandom(country string)

client.ResetIteration()
```

**Example**

```go
package main

import (
	"fmt"
	"github.com/grzfrmbl/mullCtrl"
	"log"
	"time"
)

func main() {
	client := mullCtrl.NewMullControlClient()
	for i := 0; i < 4; i++ {
		err := client.IterateAllRandom()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second*3)
		status, err := client.GetStatus()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(status.Country)
	}
}
```
