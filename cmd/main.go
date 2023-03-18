package main

import email "github.com/more-than-code/messaging"

func main() {
	port := 8002

	email.NewServer(port)
}
