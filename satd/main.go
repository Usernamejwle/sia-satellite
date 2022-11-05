package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

func getAPIPassword() string {
	apiPassword := os.Getenv("SATD_API_PASSWORD")
	if apiPassword != "" {
		fmt.Println("Using SATD_API_PASSWORD environment variable.")
	} else {
		fmt.Print("Enter API password: ")
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}
		apiPassword = string(pw)
	}
	return apiPassword
}

func main() {
	// Parse command line flags.
	log.SetFlags(0)
	userAgent := flag.String("agent", "Sat-Agent", "custom agent used for API calls")
	gatewayAddr := flag.String("addr", ":0", "address to listen on for peer connections")
	apiAddr := flag.String("api-addr", "localhost:9980", "address to serve API on")
	//satelliteAddr := flag.String("sat-addr", ":9999", "address to listen on for renter requests")
	dir := flag.String("dir", ".", "directory to store node state in")
	bootstrap := flag.Bool("bootstrap", true, "bootstrap the gateway and consensus modules")
	flag.Parse()

	// Fetch API password.
	apiPassword := getAPIPassword()

	// Start satd. startDaemon will only return when it is shutting down.
	err := startDaemon(*userAgent, *gatewayAddr, *apiAddr, apiPassword, *dir, *bootstrap)
	if err != nil {
		log.Fatal(err)
	}

	// Daemon seems to have closed cleanly. Print a 'closed' message.
	fmt.Println("Shutdown complete.")
}
