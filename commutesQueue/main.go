package main

import (
	"encoding/json"
	"fmt"
	"github.com/supabase-community/supabase-go"
	"os"
)

var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")

type Route struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	Active         bool   `json:"active"`
	StartAddress   string `json:"start_address"`
	EndAddress     string `json:"end_address"`
	StartLatitude  string `json:"start_latitude"`
	StartLongitude string `json"start_longitude"`
	StopLatitude   string `json:"stop_latitude"`
	StopLongitude  string `json:"stop_longitude"`
}

func main() {
	databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initalize client", err)
		return
	}
	response, rowsReturned, err := databaseClient.From("routes").Select("*", "exact", false).Eq("active", "true").Execute()
	if err != nil {
		fmt.Printf("Failed to insert data: %v", err)
		return
		// return Response{}, fmt.Errorf("failed to insert data: %v", err)
	}

	if rowsReturned <= 0 {
		fmt.Printf("No rows returned: %d", rowsReturned)
		return
		// return Response{}, fmt.Errorf("incorrect of rows effected, rows effected: %d", rowsEffected)
	}

	var routes []Route
	err = json.Unmarshal(response, &routes)
	if err != nil {
		fmt.Printf("Error unmarshalling response: %v", err)
	}

	for _, route := range routes {
		fmt.Printf("ID: %d, Name: %s, Active: %t\n", route.ID, route.StartAddress, route.Active)
	}
	fmt.Println("Query Succesful:")
}
