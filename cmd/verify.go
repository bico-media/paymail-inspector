package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verifies if a paymail is associated to a pubkey",
	Long: chalk.Green.Color(`
                   .__  _____       
___  __ ___________|__|/ ____\__.__.
\  \/ // __ \_  __ \  \   __<   |  |
 \   /\  ___/|  | \/  ||  |  \___  |
  \_/  \___  >__|  |__||__|  / ____|
           \/                \/`) + `
` + chalk.Yellow.Color(`
Verify will check the paymail address against a given pubkey using the provider domain (if capability is supported).

This capability allows clients to verify if a given public key is a valid identity key for a given paymail handle.

The public key returned by pki flow for a given paymail handle may change over time. 
This situation may produce troubles to verify data signed using old keys, because even having the keys, 
the verifier doesn't know if the public key actually belongs to the right user.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/05-verify-public-key-owner.html")),
	Aliases:    []string{"verification"},
	SuggestFor: []string{"pubkey"},
	Example:    configDefault + " verify mrz@" + defaultDomainName + " 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return chalker.Error("verify requires a paymail address AND pubkey")
		} else if len(args) > 2 {
			return chalker.Error("verify only supports one address and one pubkey at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		var paymailAddress string
		var pubKey string

		// Check for paymail in both args
		if strings.Contains(args[0], "@") {
			paymailAddress = args[0]
			pubKey = args[1]
		} else if strings.Contains(args[1], "@") {
			pubKey = args[0]
			paymailAddress = args[1]
		}

		// Require a paymail address
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("One argument must be a paymail address [%s] [%s]", args[0], args[1]))
			return
		}

		// Require a pubkey
		if len(pubKey) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("One argument must be a pubkey [%s] [%s]", args[0], args[1]))
			return
		}

		// Extract the parts given
		domain, _ := paymail.ExtractParts(paymailAddress)

		// Validate the paymail address and domain (error already shown)
		if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
			return
		}

		// Validate pubkey
		if len(pubKey) != paymail.PubKeyLength {
			chalker.Log(chalker.ERROR, fmt.Sprintf("PubKey is an invalid length, expected: %d but got: %d", paymail.PubKeyLength, len(pubKey)))
			return
		}

		// Get the capabilities
		capabilities, err := getCapabilities(domain)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		// Set the URL - Does the paymail provider have the capability?
		url := capabilities.GetValueString(paymail.BRFCVerifyPublicKeyOwner, "")
		if len(url) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("The provider %s is missing a required capability: %s", domain, paymail.BRFCVerifyPublicKeyOwner))
			return
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Fire the verify request
		var verify *paymail.VerifyPubKeyResponse
		if verify, err = verifyPubKey(url, parts[0], domain, pubKey); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("verify pubkey request failed: %s", err.Error()))
			return
		}

		// Rendering profile information
		displayHeader(chalker.DEFAULT, fmt.Sprintf("Rendering verify response for %s...", chalk.Cyan.Color(paymailAddress)))

		// Show the results
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("Paymail: %s", chalk.Cyan.Color(paymailAddress)))
		chalker.Log(chalker.DEFAULT, fmt.Sprintf("PubKey : %s", chalk.Cyan.Color(pubKey)))

		if verify.Match {
			chalker.Log(chalker.SUCCESS, "Paymail & PubKey Match! (service responded: match=true)")
		} else {
			chalker.Log(chalker.ERROR, "DO NOT MATCH! (service responded: match=false)")
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
