package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/supabase-community/supabase-go"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	UserID      *int    `json:"user_id" validate:"required"`
	Origin      *string `json:"origin_address" validate:"required"`
	Destination *string `json:"destination_address" validate:"required"`
	Timezone    *string `json:"timezone" validate:"required"`
	RouteSchedule
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	AddedPlaces AddedPlaces `json:"added_routes"`
}

type Place struct {
	LatLng  Coordinates `json:"coordinates"`
	Address string      `json:"address"`
}

type AddedPlaces struct {
	Origin      Place `json:"origin"`
	Destination Place `json:"destination"`
}

type Coordinates struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Geometry struct {
	Location Location `json:"location"`
}

type GeoCodeResult struct {
	Geometry         Geometry `json:"geometry"`
	FormattedAddress string   `json:"formatted_address"`
}

type GoogleResponse struct {
	GeoCodeResults []GeoCodeResult `json:"results"`
}

type InsertResponse struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	StartAddress   string `json:"start_address"`
	EndAddress     string `json:"end_address"`
	StartLatitude  string `json:"start_latitude"`
	EndLatitude    string `json:"end_latitude"`
	Active         bool   `json:"active"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	StartLongitude string `json:"start_longitude"`
	EndLongitude   string `json:"end_longitude"`
	TimeZone       string `json:"time_zone"`
}

type Route struct {
	UserID               int    `json:"user_id"`
	Origin               string `json:"start_address"`
	Destination          string `json:"end_address"`
	Timezone             string `json:"time_zone"`
	Active               bool   `json:"active"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	OriginLatitude       string `json:"start_latitude"`
	OriginLongitude      string `json:"start_longitude"`
	DestinationLatitude  string `json:"end_latitude"`
	DestinationLongitude string `json:"end_longitude"`
}

type RouteSchedule struct {
	MorningStartTime   	string `json:"morning_start_time" validate:"required,validTimeFormat"`
	MorningEndTime     	string `json:"morning_end_time" validate:"required,validTimeFormat"`
	AfternoonStartTime 	string `json:"afternoon_start_time" validate:"required,validTimeFormat"`
	AfternoonEndTime   	string `json:"afternoon_end_time" validate:"required,validTimeFormat"`
	RouteID				int		`json:"route_id"`
	Monday				bool	`json:"monday"`
	Tuesday				bool	`json:"tuesday"`
	Wednesday			bool	`json:"wednesday"`
	Thursday			bool	`json:"thursday"`
	Friday				bool	`json:"friday"`
	Saturday			bool	`json:"saturday"`
	Sunday				bool	`json:"sunday"`
}

var googleMapsAPIKey = os.Getenv("GOOGLE_API_KEY")
var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseKey = os.Getenv("SUPABASE_KEY")
var endDateBuffer = 30 //30 days from now route will become inactive

func validTimeFormat(fl validator.FieldLevel) bool {
	// Parse the time in the format "HH:mm:ss"
	_, err := time.Parse("15:04:05", fl.Field().String())
	return err == nil
}

func getCoordinates(address *string) (Place, error) {
	if address == nil {
		return Place{}, fmt.Errorf("invalid address, address not included or nil")
	}
	formattedString := *address
	// Replace commas with plus signs
	formattedString = strings.ReplaceAll(formattedString, ",", "+")
	// Replace spaces with plus signs
	formattedString = strings.ReplaceAll(formattedString, " ", "+")
	getUrl := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", formattedString, googleMapsAPIKey)
	response, err := http.Get(getUrl)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return Place{}, fmt.Errorf("error sending GET request: %s", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return Place{}, fmt.Errorf("error reading response body: %s", err)
	}

	var apiResponse GoogleResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Place{}, fmt.Errorf("error unmarshalling JSON: %s", err)
	}
	// Verify length of results
	if len(apiResponse.GeoCodeResults) != 1 {
		fmt.Println("Expected 1 result, got:", len(apiResponse.GeoCodeResults))
		return Place{}, fmt.Errorf("expected 1 result, got: %d", len(apiResponse.GeoCodeResults))
	}
	// Check if geometry.location.lat and geometry.location.lng are present
	firstResult := apiResponse.GeoCodeResults[0]
	if firstResult.Geometry.Location.Lat == 0 || firstResult.Geometry.Location.Lng == 0 {
		fmt.Println("Latitude or Longitude is missing or zero.")
		return Place{}, fmt.Errorf("latitude or Longitude is missing or zero")
	}
	resultingPlace := Place{
		LatLng: Coordinates{
			Latitude:  strconv.FormatFloat(firstResult.Geometry.Location.Lat, 'f', -1, 64),
			Longitude: strconv.FormatFloat(firstResult.Geometry.Location.Lng, 'f', -1, 64),
		},
		Address: firstResult.FormattedAddress,
	}
	return resultingPlace, nil
}

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	if googleMapsAPIKey == ""{
		return Response{}, fmt.Errorf("error loading google maps API key from environment variables")
	}
	validate := validator.New()
	validate.RegisterValidation("validTimeFormat", validTimeFormat)

	if err := validate.Struct(request); err != nil {
        for _, err := range err.(validator.ValidationErrors) {
            fmt.Printf("Validation failed for field '%s': %s\n", err.Field(), err.Tag())
			return Response{}, fmt.Errorf("validation failed for field '%s': %s", err.Field(), err.Tag())
        }
    }
	// Validate timezone string
	userLocation, err := time.LoadLocation(*request.Timezone)
	if err != nil {
		return Response{}, fmt.Errorf("invalid timezone: %s", request.Timezone)
	}
	databaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initialize client", err)
	}
	originPlace, err := getCoordinates(request.Origin)
	if err != nil {
		fmt.Println("Error obtaining origin location coordinates:", err)
		return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
	}
	destinationPlace, err := getCoordinates(request.Destination)
	if err != nil {
		fmt.Println("Error obtaining destination location coordinates:", err)
		return Response{}, fmt.Errorf("error obtaining location coordinates: %s", err)
	}
	newRoute := Route{
		UserID:               *request.UserID,
		Origin:               originPlace.Address,
		OriginLatitude:       originPlace.LatLng.Latitude,
		OriginLongitude:      originPlace.LatLng.Longitude,
		Destination:          destinationPlace.Address,
		DestinationLatitude:  destinationPlace.LatLng.Latitude,
		DestinationLongitude: destinationPlace.LatLng.Longitude,
		Timezone:             *request.Timezone,
		Active:               true,
		StartDate:            time.Now().In(userLocation).Format("2006-01-02"),
		EndDate:              time.Now().In(userLocation).AddDate(0, 0, endDateBuffer).Format("2006-01-02"),
	}
	response, rowsEffected, err := databaseClient.From("routes").Insert(newRoute, true, "", "", "exact").Execute()
	if err != nil {
		fmt.Printf("Failed to insert data: %v", err)
		return Response{}, fmt.Errorf("failed to insert data: %v", err)
	}

	if rowsEffected != 1 {
		fmt.Printf("Incorrect of rows affected, rows affected: %d", rowsEffected)
		return Response{}, fmt.Errorf("incorrect of rows affected, rows affected: %d", rowsEffected)
	}

	var insertedRows []InsertResponse
	err = json.Unmarshal(response, &insertedRows)
    if err != nil {
        return Response{}, fmt.Errorf("failed to unmarshal route insert response: %v", err)
    }

	fmt.Printf("Routes Response: %s\n", string(response))
	addUserRouteResponse := Response{
		Message: "Success",
		Data: Data{
			AddedPlaces: AddedPlaces{
				Origin:      originPlace,
				Destination: destinationPlace,
			},
		},
	}
	routeID := insertedRows[0].ID
	request.RouteSchedule.RouteID = int(routeID)
	scheduleQuery := databaseClient.From("route_schedule").Insert(request.RouteSchedule, true, "", "", "exact")
    response, rowsEffected, err = scheduleQuery.Execute()
    if err != nil {
        return Response{}, fmt.Errorf("failed to insert into routes_schedule: %v", err) 
    }

	if rowsEffected != 1 {
		fmt.Printf("Incorrect number of rows affected, rows affected: %d", rowsEffected)
		return Response{}, fmt.Errorf("incorrect of number rows affected, rows affected: %d", rowsEffected)
	}
	fmt.Printf("Routes Schedule Response: %s\n", string(response))
	return addUserRouteResponse, nil
}

func getTimeInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	timeInput, _ := reader.ReadString('\n')

	// Trim any leading/trailing whitespace
	timeInput = strings.TrimSpace(timeInput)

	// Parse the time input to a time.Time object
	parsedTime, err := time.Parse("03:04 PM", timeInput)
	if err != nil {
		// If the input is invalid, print an error message and return an empty string
		fmt.Println("Invalid time format. Please use the format 12:34 PM.")
		return ""
	}
	return parsedTime.Format("15:04:05")
}

func getBooleanInput(prompt string) bool {
	var input string
	fmt.Print(prompt + " (y/n): ")
	fmt.Scanln(&input)
	return strings.ToLower(input) == "y"
}

func main() {
	// Prompt for input from the user
	var userID int
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter User ID: ")
	fmt.Scanln(&userID)

	fmt.Print("Enter Origin Address: ")
	origin, _ := reader.ReadString('\n')
	origin = origin[:len(origin)-1]

	fmt.Print("Enter Destination Address: ")
	destination, _ := reader.ReadString('\n')
	destination = destination[:len(destination)-1]

	validTimezones := []string{
		"UTC", "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", "Europe/London",
		"Europe/Paris", "Asia/Tokyo", "Asia/Kolkata", "Australia/Sydney", "Africa/Johannesburg",
		"Asia/Singapore", "Europe/Berlin", "America/Mexico_City", "Europe/Rome", "Asia/Dubai",
	}

	fmt.Println("Please choose a timezone from the list:")
	for i, tz := range validTimezones {
		fmt.Printf("%d: %s\n", i+1, tz)
	}

	fmt.Print("Enter the number corresponding to your timezone: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input) // Remove any trailing newline or space

	// Convert the input to an integer and validate
	var timezone string
	if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(validTimezones) {
		timezone = validTimezones[idx-1]
	} else {
		fmt.Println("Invalid selection. Please choose a valid timezone number.")
		return
	}

	morningStart := getTimeInput("Enter Morning Start Time (12-hour format, e.g., 08:30 PM): ")
	morningEnd := getTimeInput("Enter Morning End Time (12-hour format, e.g., 08:30 PM): ")
	afternoonStart := getTimeInput("Enter Afternoon Start Time (12-hour format, e.g., 08:30 PM): ")
	afternoonEnd := getTimeInput("Enter Afternoon End Time (12-hour format, e.g., 08:30 PM): ")

	monday := getBooleanInput("Is this schedule active on Monday?")
	tuesday := getBooleanInput("Is this schedule active on Tuesday?")
	wednesday := getBooleanInput("Is this schedule active on Wednesday?")
	thursday := getBooleanInput("Is this schedule active on Thursday?")
	friday := getBooleanInput("Is this schedule active on Friday?")
	saturday := getBooleanInput("Is this schedule active on Saturday?")
	sunday := getBooleanInput("Is this schedule active on Sunday?")

	// Prepare the request
	request := Request{
		UserID:      &userID,
		Origin:      &origin,
		Destination: &destination,
		Timezone:    &timezone,
		RouteSchedule: RouteSchedule{
			MorningStartTime:   morningStart,
			MorningEndTime:     morningEnd,
			AfternoonStartTime: afternoonStart,
			AfternoonEndTime:   afternoonEnd,
			Monday:             monday,
			Tuesday:            tuesday,
			Wednesday:          wednesday,
			Thursday:           thursday,
			Friday:             friday,
			Saturday:           saturday,
			Sunday:             sunday,
		},
	}

	// Handle the request
	ctx := context.Background()
	response, err := HandleRequest(ctx, request)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response
	fmt.Printf("Response: %+v\n", response)
}
