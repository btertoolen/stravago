package stravago

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Activity struct {
	ResourceState              int       `json:"resource_state"`
	Athlete                    Athlete   `json:"athlete"`
	Name                       string    `json:"name"`
	Distance                   float64   `json:"distance"`
	MovingTime                 int       `json:"moving_time"`
	ElapsedTime                int       `json:"elapsed_time"`
	TotalElevationGain         float64   `json:"total_elevation_gain"`
	Type                       string    `json:"type"`
	SportType                  string    `json:"sport_type"`
	ID                         int       `json:"id"`
	StartDate                  time.Time `json:"start_date"`
	StartDateLocal             time.Time `json:"start_date_local"`
	Timezone                   string    `json:"timezone"`
	UTCOffset                  float64   `json:"utc_offset"`
	LocationCity               string    `json:"location_city"`
	LocationState              string    `json:"location_state"`
	LocationCountry            string    `json:"location_country"`
	AchievementCount           int       `json:"achievement_count"`
	KudosCount                 int       `json:"kudos_count"`
	CommentCount               int       `json:"comment_count"`
	AthleteCount               int       `json:"athlete_count"`
	PhotoCount                 int       `json:"photo_count"`
	Map                        Map       `json:"map"`
	Trainer                    bool      `json:"trainer"`
	Commute                    bool      `json:"commute"`
	Manual                     bool      `json:"manual"`
	Private                    bool      `json:"private"`
	Visibility                 string    `json:"visibility"`
	Flagged                    bool      `json:"flagged"`
	GearID                     string    `json:"gear_id"`
	StartLatlng                []float64 `json:"start_latlng"`
	EndLatlng                  []float64 `json:"end_latlng"`
	AverageSpeed               float64   `json:"average_speed"`
	MaxSpeed                   float64   `json:"max_speed"`
	AverageCadence             float64   `json:"average_cadence"`
	HasHeartrate               bool      `json:"has_heartrate"`
	AverageHeartrate           float64   `json:"average_heartrate"`
	MaxHeartrate               float64   `json:"max_heartrate"`
	HeartrateOptOut            bool      `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption bool      `json:"display_hide_heartrate_option"`
	ElevHigh                   float64   `json:"elev_high"`
	ElevLow                    float64   `json:"elev_low"`
	UploadID                   int       `json:"upload_id"`
	UploadIDStr                string    `json:"upload_id_str"`
	ExternalID                 string    `json:"external_id"`
	FromAcceptedTag            bool      `json:"from_accepted_tag"`
	PRCount                    int       `json:"pr_count"`
	TotalPhotoCount            int       `json:"total_photo_count"`
	HasKudoed                  bool      `json:"has_kudoed"`
}

type Athlete struct {
	ID            int `json:"id"`
	ResourceState int `json:"resource_state"`
}

type Map struct {
	ID              string `json:"id"`
	SummaryPolyline string `json:"summary_polyline"`
	ResourceState   int    `json:"resource_state"`
}

const api_base_url = "https://www.strava.com/api/v3"
const get_athlete_url = "athletes"
const athlete_id = 130630970
const bearer = "Bearer 7e3d3f20c0bf98f7556e571b6dc61488e72d0b72"

func Hello() {
	log.Println("Hello from stravago!")
}

func Authenticate() {
}

func doRequest(request http.Request) (response string) {
	request.Header.Add("Authorization", bearer)
	res, err := http.DefaultClient.Do(&request)
	if err != nil {
		log.Println(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	response = string(resBody)
	return response
}

// Return json object in this function
func GetAthleteStats() (response string) {
	request_url := fmt.Sprintf("%s/%s/%d/stats", api_base_url, get_athlete_url, athlete_id)
	println(request_url)
	req, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		log.Println(err)
	}
	return doRequest(*req)
}

func GetAthleteRunningStats() []byte {
	allSports := GetAthleteStats()
	// Get only running stats
	var data map[string]interface{}
	err := json.Unmarshal([]byte(allSports), &data)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return nil
	}

	runningStats := data["all_run_totals"]
	fmt.Printf("running stats: %v\n", runningStats)
	result, err := json.Marshal(runningStats)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return nil
	}
	return result
}

func GetAthleteActivities() string {
	request_url := fmt.Sprintf("%s/%s/%d/activities", api_base_url, get_athlete_url, athlete_id)
	println(request_url)
	req, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		log.Println(err)
	}
	return doRequest(*req)
}

func GetWeeklyVolumes() string {
	input := GetAthleteActivities()
	// Get the date of the first upcoming Sunday from now
	now := time.Now()
	daysUntilSunday := time.Sunday - now.Weekday()
	if daysUntilSunday < 0 {
		daysUntilSunday += 7
	}
	firstSunday := now.AddDate(0, 0, int(daysUntilSunday))
	monday := firstSunday.AddDate(0, 0, -6)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 1, 0, 0, monday.Location())
	var data []Activity
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
	}

	result_string := "["

	// Loop over every activity and check if it is within the week
	// starting from the upcoming Sunday
	weeklyVolume := 0.0
	// Loop over all entries in the list in data
	for _, activity := range data {
		// If the activity is within the week, add the distance to the weekly volume
		if activity.StartDate.After(monday) {
			weeklyVolume += activity.Distance
		} else {
			// If the activity is older than the week, we should save the weekly volume
			// and start a new week
			result_string += fmt.Sprintf("{\"weekStart\": \"%s\", \"volume\": %f},", monday, weeklyVolume)
			weeklyVolume = activity.Distance
			// Move the week to the previous week
			monday = monday.AddDate(0, 0, -7)
		}
	}
	// Save the last week
	result_string += fmt.Sprintf("{\"weekStart\": \"%s\", \"volume\": %f}", monday, weeklyVolume)
	result_string += "]"

	return string(result_string)
}
