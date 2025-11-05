package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func main() {
	// Hardcoded SMTP credentials
	smtpHost := "smtp.gmail.com"
	smtpPort := 465
	smtpUsername := "kirub.hel@gmail.com"
	smtpPassword := "dqys bnjk hhny khbk" // Try with spaces
	// smtpPassword := "dqysbnjkhhnykhbk" // Try without spaces if above fails
	fromEmail := "kirub.hel@gmail.com"
	toEmail := "kirubel.tech23@gmail.com"

	fmt.Printf("Testing SMTP Connection...\n")
	fmt.Printf("Host: %s\n", smtpHost)
	fmt.Printf("Port: %d\n", smtpPort)
	fmt.Printf("Username: %s\n", smtpUsername)
	fmt.Printf("Password: %s (length: %d)\n", smtpPassword, len(smtpPassword))
	fmt.Printf("From: %s\n", fromEmail)
	fmt.Printf("To: %s\n", toEmail)
	fmt.Println()

	// Create email message
	subject := "Test Email from TucanBIT"
	body := "This is a test email to verify SMTP configuration."
	message := fmt.Sprintf("From: %s\r\n", fromEmail) +
		fmt.Sprintf("To: %s\r\n", toEmail) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"\r\n" +
		body

	// TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	fmt.Printf("Connecting to %s...\n", addr)

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		fmt.Printf("ERROR: Failed to connect to SMTP server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Connected to SMTP server")

	// Create SMTP client
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		fmt.Printf("ERROR: Failed to create SMTP client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	fmt.Println("✓ SMTP client created")

	// Authenticate
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	fmt.Println("Attempting authentication...")
	if err := client.Auth(auth); err != nil {
		fmt.Printf("ERROR: Authentication failed: %v\n", err)
		fmt.Printf("\nTroubleshooting:\n")
		fmt.Printf("1. Verify the app password is correct: %s\n", smtpPassword)
		fmt.Printf("2. Check if 2-Step Verification is enabled on the Gmail account\n")
		fmt.Printf("3. Verify the app password was generated for 'Mail' application\n")
		fmt.Printf("4. Try removing spaces from the password\n")
		os.Exit(1)
	}
	fmt.Println("✓ Authentication successful")

	// Set sender
	if err := client.Mail(fromEmail); err != nil {
		fmt.Printf("ERROR: Failed to set sender: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Sender set")

	// Set recipient
	if err := client.Rcpt(toEmail); err != nil {
		fmt.Printf("ERROR: Failed to set recipient: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Recipient set")

	// Send email
	writer, err := client.Data()
	if err != nil {
		fmt.Printf("ERROR: Failed to get data writer: %v\n", err)
		os.Exit(1)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		fmt.Printf("ERROR: Failed to write message: %v\n", err)
		os.Exit(1)
	}

	err = writer.Close()
	if err != nil {
		fmt.Printf("ERROR: Failed to close writer: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Email sent successfully")

	// Quit
	client.Quit()
	fmt.Println("✓ Connection closed")
	fmt.Printf("\n✅ SUCCESS! Test email sent to %s\n", toEmail)
	fmt.Println("Check your inbox for the test email.")
}

