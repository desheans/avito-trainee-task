package main

import (
	"log"
	"os"

	"avito-trainee-task/internal/migration"
)

func main() {
	url, ok := os.LookupEnv("PG_URL")
	if !ok {
		log.Fatal("No PG_URL env variable provided")
	}

	if err := migration.Migrate("/migrations", url); err != nil {
		log.Fatalf("cmd.migrator.main. Failed migrate: %s", err.Error())
	}

	log.Println("Migrations applied")
}
