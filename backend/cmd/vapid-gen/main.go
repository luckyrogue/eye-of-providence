package main

import (
	"fmt"
	"os"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("# Add these to .env (NEVER commit private key):")
	fmt.Printf("EOP_VAPID_PUBLIC_KEY=%s\n", pub)
	fmt.Printf("EOP_VAPID_PRIVATE_KEY=%s\n", priv)
	fmt.Println("EOP_VAPID_SUBJECT=mailto:admin@your-domain.example")
}
