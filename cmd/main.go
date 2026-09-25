package main

import (
	"context"
	"fmt"
	"log"

	"encoding/json"

	"github.com/eitanoid/habit-tracker/internal/repository"
	"github.com/eitanoid/habit-tracker/internal/schemas"
)

func main() {

	ctx := context.Background()

	req := schemas.NewSchemaRequest("Step Tracking", "Number of steps taken in a walk").WithField(
		"Step count", "Number of steps taken", schemas.TypeInt, true,
	).WithField(
		"Left at", "Time of departing for the walk", schemas.TypeTimestamp, true,
	)
	s, err := req.Create()
	if err != nil {
		fmt.Printf("failed to create habit schema from request: %s", err.Error())
	}
	bytes, _ := json.Marshal(s)
	fmt.Println(string(bytes))

	repo, err := repository.NewSQLiteClient("my.db")
	if err != nil {
		log.Fatalf("failed to create database: %s", err.Error())
	}
	defer repo.Close()

	if err := repo.InsertSchema(ctx, s); err != nil {
		log.Fatalf("failed to insert schema: %s", err.Error())
	}
}
