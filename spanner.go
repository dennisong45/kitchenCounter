package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"google.golang.org/api/iterator"
	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

const (
	projectID  = "your-project-id"
	instanceID = "your-instance-id"
	databaseID = "your-database-id"
)

func main() {
	ctx := context.Background()

	// Create database client
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create database admin client: %v", err)
	}
	defer adminClient.Close()

	// Create the database if it doesn't exist
	op, err := adminClient.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
		CreateStatement: "CREATE DATABASE `" + databaseID + "`",
		ExtraStatements: []string{
			`CREATE TABLE Users (
				UserID   STRING(36) NOT NULL,
				Name     STRING(MAX) NOT NULL,
				Email    STRING(MAX)
			) PRIMARY KEY (UserID)`,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	if _, err := op.Wait(ctx); err != nil {
		log.Fatalf("Failed to wait for database creation: %v", err)
	}

	// Create Spanner client
	dataClient, err := spanner.NewClient(ctx, fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID))
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer dataClient.Close()

	// Insert data
	_, err = dataClient.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT Users (UserID, Name, Email) VALUES
				  (@userId, @name, @email)`,
			Params: map[string]interface{}{
				"userId": "user1",
				"name":   "John Doe",
				"email":  "john@example.com",
			},
		}
		_, err := txn.Update(ctx, stmt)
		return err
	})
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	// Query data
	stmt := spanner.Statement{SQL: `SELECT UserID, Name, Email FROM Users`}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating through results: %v", err)
		}

		var userID, name, email string
		if err := row.Columns(&userID, &name, &email); err != nil {
			log.Fatalf("Failed to parse row: %v", err)
		}
		fmt.Printf("User: %s, Name: %s, Email: %s\n", userID, name, email)
	}
}
