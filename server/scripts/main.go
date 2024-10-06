package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	var migrationName string

	if migrationName == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter migration name: ")
		migrationNameInput, _ := reader.ReadString('\n')
		migrationName = strings.TrimSpace(migrationNameInput)
		migrationName = strings.ReplaceAll(migrationName, " ", "_")
	}

	generateMigration(migrationName)
}

func generateMigration(migrationName string) {
	timestamp := time.Now().UTC().UnixMilli()

	upFile := fmt.Sprintf("%d_%s.up.sql", timestamp, migrationName)
	downFile := fmt.Sprintf("%d_%s.down.sql", timestamp, migrationName)

	migrationDir := "db/migrations"
	upFilePath := fmt.Sprintf("%s/%s", migrationDir, upFile)
	downFilePath := fmt.Sprintf("%s/%s", migrationDir, downFile)

	up, err := os.Create(upFilePath)
	if err != nil {
		log.Fatalf("Failed to create UP migration. Error: %v", err)
	}

	defer up.Close()

	down, err := os.Create(downFilePath)
	if err != nil {
		log.Fatalf("Failed to create DOWN migration. Error: %v", err)
	}

	defer down.Close()

	fmt.Println("Migration files created successfully.")
}
