package main

import (
	"fmt"
	"os"

	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Hello())
	fmt.Println(os.Args)
}
