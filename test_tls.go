package main

import (
	"fmt"

	tls_client "github.com/bogdanfinn/tls-client"
)

func main() {
	var _ tls_client.HttpClient
	fmt.Println("Success")
}
