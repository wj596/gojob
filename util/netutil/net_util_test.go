package netutil

import (
	"fmt"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	fmt.Println(GetFreePort("127.0.0.1"))
}
