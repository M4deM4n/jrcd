package main

import (
	"fmt"
)

func handleError(err error) {
	// todo: handle errors how ever I decide to.
	fmt.Println(err.Error())
}
