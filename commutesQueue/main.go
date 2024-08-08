package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awslambda "github.com/aws/aws-sdk-go/service/lambda"
	_ "github.com/lib/pq"
	"os"
	"strconv"
	"sync"
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
	StartLatitude  string `json:"start_latitude"`
	StartLongitude string `json:"start_longitude"`
	StopLatitude   string `json:"end_latitude"`
	StopLongitude  string `json:"end_longitude"`
	Timezone       string `json:"time_zone"`
	ToWork         bool   `json:"to_work"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type OptimizeRouteRequest struct {
	UserID      int      `json:"user_id"`
	Route       int      `json:"route"`
	ToWork      bool     `json:"to_work"`
	Timezone    string   `json:"timezone"`
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
}

var supabaseUsername = os.Getenv("SUPABASE_USERNAME")
var supabasePassword = os.Getenv("SUPABASE_PASSWORD")
var supabaseHost = os.Getenv("SUPABASE_HOST")
var supabasePort = os.Getenv("SUPABASE_PORT")
var supabaseDatabase = os.Getenv("SUPABASE_DATABASE")
var toWorkQueryFilePath = "to_work_valid_rows.sql"
var fromWorkQueryFilePath = "from_work_valid_rows.sql"

func processRoute(route Route, svc *awslambda.Lambda, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	floatStartLatitude, err := strconv.ParseFloat(route.StartLatitude, 64)
	if err != nil {
		resultChan <- fmt.Sprintf("Error converting starting latitude to float: %v", err)
		return
	}
	floatStartLongitude, err := strconv.ParseFloat(route.StartLongitude, 64)
	if err != nil {
		resultChan <- fmt.Sprintf("Error converting starting longitude to float: %v", err)
		return
	}
	floatDestinationLatitude, err := strconv.ParseFloat(route.StopLatitude, 64)
	if err != nil {
		resultChan <- fmt.Sprintf("Error converting destination latitude to float: %v", err)
		return
	}
	floatDestinationLongitude, err := strconv.ParseFloat(route.StopLongitude, 64)
	if err != nil {
		resultChan <- fmt.Sprintf("Error converting destination longitude to float: %v", err)
		return
	}
	routeRequest := OptimizeRouteRequest{
		UserID:   route.UserID,
		Route:    route.ID,
		ToWork:   route.ToWork,
		Timezone: route.Timezone,
		Origin: Location{
			Latitude:  floatStartLatitude,
			Longitude: floatStartLongitude,
		},
		Destination: Location{
			Latitude:  floatDestinationLatitude,
			Longitude: floatDestinationLongitude,
		},
	}
	requestData, err := json.Marshal(routeRequest)
	if err != nil {
		resultChan <- fmt.Sprintf("Error marshaling optimize route request struct: %v", err)
		return
	}

	input := &awslambda.InvokeInput{
		FunctionName: aws.String(os.Getenv("OPTIMIZE_ROUTE_FUNCTION")),
		Payload:      requestData,
	}
	result, err := svc.Invoke(input)
	if err != nil {
		resultChan <- fmt.Sprintf("Failed to invoke target function: %v", err)
		return
	}

	resultChan <- fmt.Sprintf("Successful commutes request: %v", result)
}

func fetchRoutes(query string, toWork bool, db *sql.DB, wg *sync.WaitGroup, results chan<- Route, errors chan<- error) {
	defer wg.Done()

	rows, err := db.Query(query)
	if err != nil {
		errors <- fmt.Errorf("failed to query data: %w", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var route Route
		if err := rows.Scan(
			&route.ID,
			&route.UserID,
			&route.Active,
			&route.StartLatitude,
			&route.StartLongitude,
			&route.StopLatitude,
			&route.StopLongitude,
			&route.Timezone); err != nil {
			errors <- fmt.Errorf("error scanning row: %w", err)
			continue
		}
		route.ToWork = toWork
		results <- route
	}

	if err := rows.Err(); err != nil {
		errors <- fmt.Errorf("error iterating over rows: %w", err)
	}
}

func loadQuery(filename string) (string, error) {
	query, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(query), nil
}

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	svc := awslambda.New(sess)

	toWorkQuery, err := loadQuery(toWorkQueryFilePath)
	if err != nil {
		fmt.Printf("Failed to load query: %v, at filepath: %v", err, toWorkQueryFilePath)
		return Response{}, fmt.Errorf("failed to load query: %v, at filepath: %v", err, toWorkQueryFilePath)
	}
	fromWorkQuery, err := loadQuery(fromWorkQueryFilePath)
	if err != nil {
		fmt.Printf("Failed to load query: %v, at filepath: %v", err, toWorkQueryFilePath)
		return Response{}, fmt.Errorf("failed to load query: %v, at filepath: %v", err, toWorkQueryFilePath)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", supabaseUsername, supabasePassword,
		supabaseHost, supabasePort, supabaseDatabase)

	databaseClient, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("cannot initalize database client", err)
		return Response{}, fmt.Errorf("cannot initalize database client: %v", err)
	}
	defer databaseClient.Close()

	err = databaseClient.Ping()
	if err != nil {
		fmt.Printf("Failed to ping the database: %v", err)
		return Response{}, fmt.Errorf("failed to ping the database: %v", err)
	}

	var wg sync.WaitGroup

	results := make(chan Route)
	errors := make(chan error)

	wg.Add(2)
	go func() {
		go fetchRoutes(toWorkQuery, true, databaseClient, &wg, results, errors)
		go fetchRoutes(fromWorkQuery, false, databaseClient, &wg, results, errors)
		wg.Wait()
		close(results)
		close(errors)
	}()

	var routes []Route
	for {
		select {
		case route, ok := <-results:
			if !ok {
				results = nil
			} else {
				routes = append(routes, route)
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		}
		if results == nil && errors == nil {
			break
		}
	}

	if len(routes) <= 0 {
		fmt.Printf("No rows returned: %d", len(routes))
		return Response{}, fmt.Errorf("no rows returned: %d", len(routes))
	}

	resultChan := make(chan string, len(routes))
	succeeded := 0
	failed := 0

	wg.Add(len(routes))
	for _, route := range routes {
		go processRoute(route, svc, resultChan, &wg)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		if result[:5] == "Error" {
			fmt.Println(result)
			failed++
		} else {
			succeeded++
		}
	}

	lambdaResponse := Data{
		Processed: len(routes),
		Suceeded:  succeeded,
		Failed:    failed,
	}
	return Response{"Commutes Requests Complete.", lambdaResponse}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
