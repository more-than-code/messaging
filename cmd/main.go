package main

import "github.com/more-than-code/messaging"

func main() {
	port := 8002

	messaging.NewServer(port)
}
