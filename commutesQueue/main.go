package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awslambda "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/supabase-community/supabase-go"
	"os"
	"strconv"
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
	StopLatitude   string `json:"end_latitude"`
	StopLongitude  string `json:"end_longitude"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type OptimizeRouteRequest struct {
	UserID      int      `json:"user_id"`
	Route       int      `json:"route"`
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
}

var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	svc := awslambda.New(sess)
	input := &awslambda.InvokeInput{
		FunctionName: aws.String(os.Getenv("OPTIMIZE_ROUTE_FUNCTION")),
		Payload:      []byte(`{}`),
	}
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
	suceeded := 0
	failed := 0
	totalProcessed := len(routes)
	for _, route := range routes {
		floatStartLatitude, err := strconv.ParseFloat(route.StartLatitude, 64)
		if err != nil {
			fmt.Printf("Error converting starting latitude to float: %v", err)
			return Response{}, fmt.Errorf("error converting starting latitude to float: %v", err)
		}
		floatStartLongitude, err := strconv.ParseFloat(route.StartLongitude, 64)
		if err != nil {
			fmt.Printf("Error converting starting longitude to float: %v", err)
			return Response{}, fmt.Errorf("error converting starting longitude to float: %v", err)
		}
		floatDestinationLatitude, err := strconv.ParseFloat(route.StopLatitude, 64)
		if err != nil {
			fmt.Printf("Error converting destination latitude to float: %v", err)
			return Response{}, fmt.Errorf("error converting destination latitude to float: %v", err)
		}
		floatDestinationLongitude, err := strconv.ParseFloat(route.StopLongitude, 64)
		if err != nil {
			fmt.Printf("Error converting destination longitude to float: %v", err)
			return Response{}, fmt.Errorf("error converting destination longitude to float: %v", err)
		}
		routeRequest := OptimizeRouteRequest{
			UserID: route.UserID,
			Route:  route.ID,
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
			fmt.Println("Error marshaling optimize route request struct:", err)
			return Response{}, fmt.Errorf("error marshaling optimize route request struct: %v", err)
		}
		input.Payload = requestData
		result, err := svc.Invoke(input)
		if err != nil {
			fmt.Printf("failed to invoke target function: %v", err)
			failed++
		} else {
			fmt.Println("Succesful commutes request: ", result)
			suceeded++
		}

	}
	lambdaResponse := Data{
		Processed: totalProcessed,
		Suceeded:  suceeded,
		Failed:    failed,
	}
	return Response{"Commutes Requests Complete.", lambdaResponse}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
