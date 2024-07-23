package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/supabase-community/supabase-go"
	"os"
)

type Request struct {
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Processed int `json:"routes_processed"`
	Suceeded  int `json:"routes_suceeded"`
	Failed    int `json:"routes_failed"`
}

type Route struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	Active         bool   `json:"active"`
	StartAddress   string `json:"start_address"`
	EndAddress     string `json:"end_address"`
	StartLatitude  string `json:"start_latitude"`
	StartLongitude string `json:"start_longitude"`
	StopLatitude   string `json:"stop_latitude"`
	StopLongitude  string `json:"stop_longitude"`
}

var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initalize database client", err)
		return Response{}, fmt.Errorf("cannot initalize database client: %v", err)
	}
	response, rowsReturned, err := databaseClient.From("routes").Select("*", "exact", false).Eq("active", "true").Execute()
	if err != nil {
		fmt.Printf("Failed to query data: %v", err)
		return Response{}, fmt.Errorf("failed to query data: %v", err)
	}

	if rowsReturned <= 0 {
		fmt.Printf("No rows returned: %d", rowsReturned)
		return Response{}, fmt.Errorf("no rows returned: %d", rowsReturned)
	}

	var routes []Route
	err = json.Unmarshal(response, &routes)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %v", err)
		return Response{}, fmt.Errorf("error unmarshalling response: %v", err)
	}

	for _, route := range routes {
		fmt.Printf("ID: %d, Name: %s, Active: %t\n", route.ID, route.StartAddress, route.Active)
	}
	lambdaResponse := Data{
		Processed: int(rowsReturned),
		Suceeded:  int(rowsReturned),
		Failed:    0,
	}
	return Response{"Commutes Requests Complete.", lambdaResponse}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
