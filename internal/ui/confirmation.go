package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// ConfirmationResult represents the result of a user confirmation
type ConfirmationResult int

const (
	ConfirmYes ConfirmationResult = iota
	ConfirmNo
	ConfirmCancel
)

// AskUserConfirmation prompts the user for yes/no confirmation
func AskUserConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", message)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))
		switch response {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		default:
			fmt.Printf("Please answer yes (y) or no (n).\n")
		}
	}
}

// AskRenewalConfirmation asks user about certificate renewal with context
func AskRenewalConfirmation(domains []string, daysRemaining int, expiresAt string) bool {
	if daysRemaining <= 0 {
		// Expired certificate, auto-renew
		return true
	} else if daysRemaining <= 30 {
		// Expires soon, ask user
		message := fmt.Sprintf("Certificate for %s expires in %d days (%s). Do you want to renew it now?",
			strings.Join(domains, ", "), daysRemaining, expiresAt)
		return AskUserConfirmation(message)
	} else {
		// Valid certificate, ask if user wants to force renew
		message := fmt.Sprintf("Certificate for %s is valid and expires in %d days (%s). Do you want to renew it anyway?",
			strings.Join(domains, ", "), daysRemaining, expiresAt)
		return AskUserConfirmation(message)
	}
}

// AskDomainReplaceConfirmation asks user about replacing certificate with different domains
func AskDomainReplaceConfirmation(existingDomains, newDomains []string) bool {
	fmt.Printf("⚠️  Found existing certificate with different domains:\n")
	fmt.Printf("   Existing: %s\n", strings.Join(existingDomains, ", "))
	fmt.Printf("   Requested: %s\n", strings.Join(newDomains, ", "))

	return AskUserConfirmation("Do you want to replace it with the new certificate?")
}
