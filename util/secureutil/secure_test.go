package secureutil

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestHmacMD5(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fmt.Println(timestamp)
	fmt.Println(HmacMD5("/cluster/leader"+timestamp, "gojob"))
}
