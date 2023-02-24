package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mike76-dev/sia-satellite/modules"
	"github.com/mike76-dev/sia-satellite/node/api"
	"github.com/spf13/cobra"

	"gitlab.com/NebulousLabs/encoding"
	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
	"gitlab.com/NebulousLabs/errors"

	"go.sia.tech/siad/crypto"
	smodules "go.sia.tech/siad/modules"
	"go.sia.tech/siad/modules/wallet"
	"go.sia.tech/siad/types"

	"golang.org/x/term"
)

var (
	walletAddressCmd = &cobra.Command{
		Use:   "address",
		Short: "Get a new wallet address",
		Long:  "Generate a new wallet address from the wallet's primary seed.",
		Run:   wrap(walletaddresscmd),
	}

	walletAddressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "List all addresses",
		Long:  "List all addresses that have been generated by the wallet.",
		Run:   wrap(walletaddressescmd),
	}

	walletBalanceCmd = &cobra.Command{
		Use:   "balance",
		Short: "View wallet balance",
		Long:  "View wallet balance, including confirmed and unconfirmed balance.",
		Run:   wrap(walletbalancecmd),
	}

	walletBroadcastCmd = &cobra.Command{
		Use:   "broadcast [txn]",
		Short: "Broadcast a transaction",
		Long: `Broadcast a JSON-encoded transaction to connected peers. The transaction must
be valid. txn may be either JSON, base64, or a file containing either.`,
		Run: wrap(walletbroadcastcmd),
	}

	walletChangepasswordCmd = &cobra.Command{
		Use:   "change-password",
		Short: "Change the wallet password",
		Long:  "Change the encryption password of the wallet, re-encrypting all keys + seeds kept by the wallet.",
		Run:   wrap(walletchangepasswordcmd),
	}

	walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Perform wallet actions",
		Long: `Generate a new address, send coins to another wallet, or view info about the wallet.
Units:
The smallest unit of siacoins is the hasting. One siacoin is 10^24 hastings. Other supported units are:
  pS (pico,  10^-12 SC)
  nS (nano,  10^-9 SC)
  uS (micro, 10^-6 SC)
  mS (milli, 10^-3 SC)
  SC
  KS (kilo, 10^3 SC)
  MS (mega, 10^6 SC)
  GS (giga, 10^9 SC)
  TS (tera, 10^12 SC)`,
		Run: wrap(walletbalancecmd),
	}

	walletInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize and encrypt a new wallet",
		Long: `Generate a new wallet from a randomly generated seed, and encrypt it.
By default the wallet encryption / unlock password is the same as the generated seed.`,
		Run: wrap(walletinitcmd),
	}

	walletInitSeedCmd = &cobra.Command{
		Use:   "init-seed",
		Short: "Initialize and encrypt a new wallet using a pre-existing seed",
		Long:  `Initialize and encrypt a new wallet using a pre-existing seed.`,
		Run:   wrap(walletinitseedcmd),
	}

	walletLoadCmd = &cobra.Command{
		Use:   "load",
		Short: "Load a wallet seed",
		// Run field is not set, as the load command itself is not a valid command.
		// A subcommand must be provided.
	}

	walletLoadSeedCmd = &cobra.Command{
		Use:   `seed`,
		Short: "Add a seed to the wallet",
		Long:  "Loads an auxiliary seed into the wallet.",
		Run:   wrap(walletloadseedcmd),
	}

	walletLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock the wallet",
		Long:  "Lock the wallet, preventing further use",
		Run:   wrap(walletlockcmd),
	}

	walletSeedsCmd = &cobra.Command{
		Use:   "seeds",
		Short: "View information about your seeds",
		Long:  "View your primary and auxiliary wallet seeds.",
		Run:   wrap(walletseedscmd),
	}

	walletSendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send siacoins to an address",
		Long:  "Send siacoins to an address",
		// Run field is not set, as the send command itself is not a valid command.
		// A subcommand must be provided.
	}

	walletSendSiacoinsCmd = &cobra.Command{
		Use:   "siacoins [amount] [dest]",
		Short: "Send siacoins to an address",
		Long: `Send siacoins to an address. 'dest' must be a 76-byte hexadecimal address.
'amount' can be specified in units, e.g. 1.23KS. Run 'wallet --help' for a list of units.
If no unit is supplied, hastings will be assumed.
A dynamic transaction fee is applied depending on the size of the transaction and how busy the network is.`,
		Run: wrap(walletsendsiacoinscmd),
	}

	walletSignCmd = &cobra.Command{
		Use:   "sign [txn] [tosign]",
		Short: "Sign a transaction",
		Long: `Sign a transaction. If siad is running with an unlocked wallet, the
/wallet/sign API call will be used. Otherwise, sign will prompt for the wallet
seed, and the signing key(s) will be regenerated.
txn may be either JSON, base64, or a file containing either.
tosign is an optional list of indices. Each index corresponds to a
TransactionSignature in the txn that will be filled in. If no indices are
provided, the wallet will fill in every TransactionSignature it has keys for.`,
		Run: walletsigncmd,
	}

	walletSweepCmd = &cobra.Command{
		Use:   "sweep",
		Short: "Sweep siacoins from a seed.",
		Long: `Sweep siacoins from a seed. The outputs belonging to the seed
will be sent to your wallet.`,
		Run: wrap(walletsweepcmd),
	}

	walletTransactionsCmd = &cobra.Command{
		Use:   "transactions",
		Short: "View transactions",
		Long:  "View transactions related to addresses spendable by the wallet, providing a net flow of siacoins for each transaction",
		Run:   wrap(wallettransactionscmd),
	}

	walletUnlockCmd = &cobra.Command{
		Use:   `unlock`,
		Short: "Unlock the wallet",
		Long: `Decrypt and load the wallet into memory.
Automatic unlocking is also supported via environment variable: if the
SATD_WALLET_PASSWORD environment variable is set, the unlock command will
use it instead of displaying the typical interactive prompt.`,
		Run: wrap(walletunlockcmd),
	}
)

const askPasswordText = "We need to encrypt the new data using the current wallet password, please provide: "

const (
	currentPasswordText = "Current Password: "
	newPasswordText     = "New Password: "
	confirmPasswordText = "Confirm: "
)

// For an unconfirmed Transaction, the TransactionTimestamp field is set to the
// maximum value of a uint64.
const unconfirmedTransactionTimestamp = ^uint64(0)

// passwordPrompt securely reads a password from stdin.
func passwordPrompt(prompt string) (pw string, err error) {
	fmt.Print(prompt)
	if insecureInput {
		pw, err = bufio.NewReader(os.Stdin).ReadString('\n')
	} else {
		var pwBytes []byte
		pwBytes, err = term.ReadPassword(int(syscall.Stdin))
		pw = string(pwBytes)
	}
	fmt.Println()
	if err != nil {
		err = fmt.Errorf("error reading password: %w", err)
	}
	return strings.TrimSpace(pw), err
}

// confirmPassword requests confirmation of a previously-entered password.
func confirmPassword(prev string) error {
	pw, err := passwordPrompt(confirmPasswordText)
	if err != nil {
		return err
	} else if pw != prev {
		return errors.New("passwords do not match")
	}
	return nil
}

// walletaddresscmd fetches a new address from the wallet that will be able to
// receive coins.
func walletaddresscmd() {
	addr, err := httpClient.WalletAddressGet()
	if err != nil {
		die("Could not generate new address:", err)
	}
	fmt.Printf("Created new address: %s\n", addr.Address)
}

// walletaddressescmd fetches the list of addresses that the wallet knows.
func walletaddressescmd() {
	addrs, err := httpClient.WalletAddressesGet()
	if err != nil {
		die("Failed to fetch addresses:", err)
	}
	for _, addr := range addrs.Addresses {
		fmt.Println(addr)
	}
}

// walletchangepasswordcmd changes the password of the wallet.
func walletchangepasswordcmd() {
	currentPassword, err := passwordPrompt(currentPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	newPassword, err := passwordPrompt(newPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	} else if err = confirmPassword(newPassword); err != nil {
		die(err)
	}
	err = httpClient.WalletChangePasswordPost(currentPassword, newPassword)
	if err != nil {
		die("Changing the password failed:", err)
	}
	fmt.Println("Password changed successfully.")
}

// walletinitcmd encrypts the wallet with the given password
func walletinitcmd() {
	var password string
	var err error
	if initPassword {
		password, err = passwordPrompt("Wallet password: ")
		if err != nil {
			die("Reading password failed:", err)
		} else if err = confirmPassword(password); err != nil {
			die(err)
		}
	}
	er, err := httpClient.WalletInitPost(password, initForce)
	if err != nil {
		die("Error when encrypting wallet:", err)
	}
	fmt.Printf("Recovery seed:\n%s\n\n", er.PrimarySeed)
	if initPassword {
		fmt.Printf("Wallet encrypted with given password\n")
	} else {
		fmt.Printf("Wallet encrypted with password:\n%s\n", er.PrimarySeed)
	}
}

// walletinitseedcmd initializes the wallet from a preexisting seed.
func walletinitseedcmd() {
	seed, err := passwordPrompt("Seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}
	var password string
	if initPassword {
		password, err = passwordPrompt("Wallet password: ")
		if err != nil {
			die("Reading password failed:", err)
		} else if err = confirmPassword(password); err != nil {
			die(err)
		}
	}
	err = httpClient.WalletInitSeedPost(seed, password, initForce)
	if err != nil {
		die("Could not initialize wallet from seed:", err)
	}
	if initPassword {
		fmt.Println("Wallet initialized and encrypted with given password.")
	} else {
		fmt.Println("Wallet initialized and encrypted with seed.")
	}
}

// walletloadseedcmd adds a seed to the wallet's list of seeds
func walletloadseedcmd() {
	seed, err := passwordPrompt("New seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}
	password, err := passwordPrompt(askPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	err = httpClient.WalletSeedPost(seed, password)
	if err != nil {
		die("Could not add seed:", err)
	}
	fmt.Println("Added Key")
}

// walletlockcmd locks the wallet
func walletlockcmd() {
	err := httpClient.WalletLockPost()
	if err != nil {
		die("Could not lock wallet:", err)
	}
}

// walletseedcmd returns the current seed {
func walletseedscmd() {
	seedInfo, err := httpClient.WalletSeedsGet()
	if err != nil {
		die("Error retrieving the current seed:", err)
	}
	fmt.Println("Primary Seed:")
	fmt.Println(seedInfo.PrimarySeed)
	if len(seedInfo.AllSeeds) == 1 {
		// AllSeeds includes the primary seed
		return
	}
	fmt.Println()
	fmt.Println("Auxiliary Seeds:")
	for _, seed := range seedInfo.AllSeeds {
		if seed == seedInfo.PrimarySeed {
			continue
		}
		fmt.Println() // extra newline for readability
		fmt.Println(seed)
	}
}

// walletsendsiacoinscmd sends siacoins to a destination address.
func walletsendsiacoinscmd(amount, dest string) {
	hastings, err := types.ParseCurrency(amount)
	if err != nil {
		die("Could not parse amount:", err)
	}
	var value types.Currency
	if _, err := fmt.Sscan(hastings, &value); err != nil {
		die("Failed to parse amount", err)
	}
	var hash types.UnlockHash
	if _, err := fmt.Sscan(dest, &hash); err != nil {
		die("Failed to parse destination address", err)
	}
	_, err = httpClient.WalletSiacoinsPost(value, hash, walletTxnFeeIncluded)
	if err != nil {
		die("Could not send siacoins:", err)
	}
	fmt.Printf("Sent %s hastings to %s\n", hastings, dest)
}

// walletbalancecmd retrieves and displays information about the wallet.
func walletbalancecmd() {
	status, err := httpClient.WalletGet()
	if errors.Contains(err, api.ErrAPICallNotRecognized) {
		// Assume module is not loaded if status command is not recognized.
		fmt.Printf("Wallet:\n  Status: %s\n\n", moduleNotReadyStatus)
		return
	} else if err != nil {
		die("Could not get wallet status:", err)
	}

	fees, err := httpClient.TransactionPoolFeeGet()
	if err != nil {
		die("Could not get fee estimation:", err)
	}
	encStatus := "Unencrypted"
	if status.Encrypted {
		encStatus = "Encrypted"
	}
	if !status.Unlocked {
		fmt.Printf(`Wallet status:
%v, Locked
Unlock the wallet to view balance
`, encStatus)
		return
	}

	unconfirmedBalance := status.ConfirmedSiacoinBalance.Add(status.UnconfirmedIncomingSiacoins).Sub(status.UnconfirmedOutgoingSiacoins)
	var delta string
	if unconfirmedBalance.Cmp(status.ConfirmedSiacoinBalance) >= 0 {
		delta = "+" + modules.CurrencyUnits(unconfirmedBalance.Sub(status.ConfirmedSiacoinBalance))
	} else {
		delta = "-" + modules.CurrencyUnits(status.ConfirmedSiacoinBalance.Sub(unconfirmedBalance))
	}

	fmt.Printf(`Wallet status:
%s, Unlocked
Height:              %v
Confirmed Balance:   %v
Unconfirmed Delta:   %v
Exact:               %v H
Estimated Fee:       %v / KB
`, encStatus, status.Height, modules.CurrencyUnits(status.ConfirmedSiacoinBalance), delta,
		status.ConfirmedSiacoinBalance, fees.Maximum.Mul64(1e3).HumanString())
}

// walletbroadcastcmd broadcasts a transaction.
func walletbroadcastcmd(txnStr string) {
	txn, err := parseTxn(txnStr)
	if err != nil {
		die("Could not decode transaction:", err)
	}
	err = httpClient.TransactionPoolRawPost(txn, nil)
	if err != nil {
		die("Could not broadcast transaction:", err)
	}
	fmt.Println("Transaction has been broadcast successfully")
}

// walletsweepcmd sweeps coins and funds from a seed.
func walletsweepcmd() {
	seed, err := passwordPrompt("Seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}

	swept, err := httpClient.WalletSweepPost(seed)
	if err != nil {
		die("Could not sweep seed:", err)
	}
	fmt.Printf("Swept %v from seed.\n", modules.CurrencyUnits(swept.Coins))
}

// walletsigncmd signs a transaction.
func walletsigncmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		_ = cmd.UsageFunc()(cmd)
		os.Exit(exitCodeUsage)
	}

	txn, err := parseTxn(args[0])
	if err != nil {
		die("Could not decode transaction:", err)
	}

	var toSign []crypto.Hash
	for _, arg := range args[1:] {
		index, err := strconv.ParseUint(arg, 10, 32)
		if err != nil {
			die("Invalid signature index", index, "(must be an non-negative integer)")
		} else if index >= uint64(len(txn.TransactionSignatures)) {
			die("Invalid signature index", index, "(transaction only has", len(txn.TransactionSignatures), "signatures)")
		}
		toSign = append(toSign, txn.TransactionSignatures[index].ParentID)
	}

	// Try API first.
	wspr, err := httpClient.WalletSignPost(txn, toSign)
	if err == nil {
		txn = wspr.Transaction
	} else {
		// If satd is running, but the wallet is locked, assume the user
		// wanted to sign with satd.
		if strings.Contains(err.Error(), smodules.ErrLockedWallet.Error()) {
			die("Signing via API failed: satd is running, but the wallet is locked.")
		}

		// satd is not running; fallback to offline keygen.
		walletsigncmdoffline(&txn, toSign)
	}

	if walletRawTxn {
		_, err = base64.NewEncoder(base64.StdEncoding, os.Stdout).Write(encoding.Marshal(txn))
	} else {
		err = json.NewEncoder(os.Stdout).Encode(txn)
	}
	if err != nil {
		die("failed to encode txn", err)
	}
	fmt.Println()
}

// walletsigncmdoffline is a helper for walletsigncmd that handles signing
// transactions without satd.
func walletsigncmdoffline(txn *types.Transaction, toSign []crypto.Hash) {
	fmt.Println("Enter your wallet seed to generate the signing key(s) now and sign without satd.")
	seedString, err := passwordPrompt("Seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}
	seed, err := smodules.StringToSeed(seedString, mnemonics.English)
	if err != nil {
		die("Invalid seed:", err)
	}
	// Signing via seed may take a while, since we need to regenerate
	// keys. If it takes longer than a second, print a message to assure
	// the user that this is normal.
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(time.Second):
			fmt.Println("Generating keys; this may take a few seconds...")
		case <-done:
		}
	}()
	err = wallet.SignTransaction(txn, seed, toSign, 180e3)
	if err != nil {
		die("Failed to sign transaction:", err)
	}
	close(done)
}

// wallettransactionscmd lists all of the transactions related to the wallet,
// providing a net flow of siacoins for each.
func wallettransactionscmd() {
	wtg, err := httpClient.WalletTransactionsGet(types.BlockHeight(walletStartHeight), types.BlockHeight(walletEndHeight))
	if err != nil {
		die("Could not fetch transaction history:", err)
	}
	cg, err := httpClient.ConsensusGet()
	if err != nil {
		die("Could not fetch consensus information:", err)
	}
	fmt.Println("             [timestamp]    [height]                                                   [transaction id]    [net siacoins]")
	txns := append(wtg.ConfirmedTransactions, wtg.UnconfirmedTransactions...)
	sts, err := wallet.ComputeValuedTransactions(txns, cg.Height)
	if err != nil {
		die("Could not compute valued transaction: ", err)
	}
	for _, txn := range sts {
		// Convert the siacoins to a float.
		incomingSiacoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedIncomingValue.Big(), types.SiacoinPrecision.Big()).Float64()
		outgoingSiacoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedOutgoingValue.Big(), types.SiacoinPrecision.Big()).Float64()

		// Print the results.
		if uint64(txn.ConfirmationTimestamp) != unconfirmedTransactionTimestamp {
			fmt.Println(time.Unix(int64(txn.ConfirmationTimestamp), 0).Format("2006-01-02 15:04:05-0700"))
		} else {
			fmt.Printf("             unconfirmed")
		}
		if txn.ConfirmationHeight < 1e9 {
			fmt.Printf("%12v", txn.ConfirmationHeight)
		} else {
			fmt.Printf(" unconfirmed")
		}
		fmt.Printf("%67v%15.2f SC", txn.TransactionID, incomingSiacoinsFloat - outgoingSiacoinsFloat)
	}
}

// walletunlockcmd unlocks a saved wallet.
func walletunlockcmd() {
	// Try reading from environment variable first, then fallback to
	// interactive method. Also allow overriding auto-unlock via -p.
	password := os.Getenv("SATD_WALLET_PASSWORD")
	if password != "" && !initPassword {
		fmt.Println("Using SATD_WALLET_PASSWORD environment variable")
		err := httpClient.WalletUnlockPost(password)
		if err != nil {
			fmt.Println("Automatic unlock failed!")
		} else {
			fmt.Println("Wallet unlocked")
			return
		}
	}
	password, err := passwordPrompt("Wallet password: ")
	if err != nil {
		die("Reading password failed:", err)
	}
	err = httpClient.WalletUnlockPost(password)
	if err != nil {
		die("Could not unlock wallet:", err)
	}
}
