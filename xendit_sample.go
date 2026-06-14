package main

import (
	"fmt"
	"log"

	"github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
)

func main() {
	// Initialize the SDK with your Secret API Key (Keep this secure!)
	xendit.Opt.SecretKey = "XENDIT_SECRET_KEY"

	// Define the parameters for the user's wallet top-up
	data := invoice.CreateParams{
		ExternalID:  "invoice-12345", // Unique identifier for this transaction
		Amount:      10.00, // Set your required amount
		Description: "Wallet Top-Up for Unicard", // Description for the transaction
		PayerEmail:  "unicard@dev.zeeh",
		Currency:    "PHP", // Adjust currency to your local demographic
	}

	// Generate the invoice via Xendit API
	resp, err := invoice.Create(&data)
	if err != nil {
		log.Fatalf("Failed to create Xendit invoice: %v", err)
	}

	// Output the secure checkout URL for the user
	fmt.Printf("Invoice created successfully!\n")
	fmt.Printf("Direct the user to this URL to complete their top-up: %s\n", resp.InvoiceURL)
}
