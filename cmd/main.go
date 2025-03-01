package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/gagliardetto/solana-go"
)

var CONTRACT_ADDRESSES = []string{
	"AHxE3UAjMzmVqWv7KdYUpEfXaXki163b2kHakTHhxszS",
	// Add more addresses here
}

const (
	// Replace with your paid RPC endpoint
	RPC_ENDPOINT = rpc.MainnetRPCEndpoint
	MAX_RETRIES  = 5
	BASE_DELAY   = 2 * time.Second
)

func main() {
	// Create RPC client
	c := client.NewClient(RPC_ENDPOINT)

	// Create a channel to handle errors
	errCh := make(chan error)

	// Start monitoring goroutine for each contract
	for _, addr := range CONTRACT_ADDRESSES {
		fmt.Println("Monitoring contract:", addr)
		go monitorContract(c, addr, errCh)
	}

	// Handle errors from goroutines
	for err := range errCh {
		log.Printf("Error monitoring contract: %v", err)
	}
}

func monitorContract(c *client.Client, contractAddress string, errCh chan<- error) {
	// Convert address to PublicKey
	pubKey, err := solana.PublicKeyFromBase58(contractAddress)
	if err != nil {
		errCh <- fmt.Errorf("invalid contract address %s: %w", contractAddress, err)
		return
	}

	var lastSignature string // Track last processed transaction

	for {
		// Get recent transactions for the contract
		txns, err := c.GetSignaturesForAddress(context.Background(), pubKey.String())
		if err != nil {
			errCh <- fmt.Errorf("failed to get transactions for %s: %w", contractAddress, err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Process transactions in reverse order (oldest first)
		for i := len(txns) - 1; i >= 0; i-- {
			txn := txns[i]
			if txn.Signature == lastSignature {
				break
			}

			// Check if we've found a token release event
			if isTokenReleaseEvent(c, txn.Signature) {
				log.Printf("Token release detected on contract %s: %s", contractAddress, txn.Signature)
				// Add your custom handling logic here
			}
		}

		// Update last processed signature
		if len(txns) > 0 {
			lastSignature = txns[0].Signature
		}

		time.Sleep(15 * time.Second) // Poll every 15 seconds
	}
}

func isTokenReleaseEvent(c *client.Client, signature string) bool {
	// Get transaction details
	txn, err := c.GetTransaction(context.Background(), signature)
	if err != nil {
		return false
	}

	// Check log messages for token release event
	for _, log := range txn.Meta.LogMessages {
		// Memecoins typically emit events like "TokenReleased" or similar in logs
		// You'll need to verify the exact log message format from the contract's documentation
		if containsTokenReleaseLog(log) {
			return true
		}
	}
	return false
}

// Customize this based on the contract's actual log message format
func containsTokenReleaseLog(log string) bool {
	// Example pattern - adjust based on actual contract events
	return strings.Contains(log, "TokenReleased") ||
		strings.Contains(log, "TokensMinted") ||
		strings.Contains(log, "Transfer") // Some might use SPL token transfer events
}
