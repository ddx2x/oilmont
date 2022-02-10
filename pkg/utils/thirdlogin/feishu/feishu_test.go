package feishu

import (
	"fmt"
	"testing"
)

func TestSUID(t *testing.T) {
	third := NewFeiShu()

	account, err := third.GetAccountAccessToken("code")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", account)
}
