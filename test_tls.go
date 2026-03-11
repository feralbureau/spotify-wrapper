package main
import (
    tls_client "github.com/bogdanfinn/tls-client"
    "fmt"
)
func main() {
    var _ tls_client.HttpClient
    fmt.Println("Success")
}
