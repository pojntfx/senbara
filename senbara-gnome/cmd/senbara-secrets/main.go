package main

import (
	"fmt"

	"github.com/pojntfx/senbara/senbara-gnome/assets/resources"
	"github.com/zalando/go-keyring"
)

const (
	idTokenKey = "id_token"
)

func main() {
	if err := keyring.Set(resources.AppID, idTokenKey, "example-id-token-value"); err != nil {
		panic(err)
	}

	it, err := keyring.Get(resources.AppID, idTokenKey)
	if err != nil {
		panic(err)
	}

	fmt.Println(it)
}
