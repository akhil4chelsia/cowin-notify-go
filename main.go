package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

//Chane as per search criteria
var doseNo uint16 = 1
var pincodes = []int32{683541, 683542, 683544, 683545, 683546, 683549, 683556, 683561, 683563, 683572, 683574, 683575, 683576, 683577, 683580, 683581, 683587, 686661, 686668, 682316, 686670, 686672, 686673, 683101, 683105, 683112}

var wg sync.WaitGroup

func main() {

	channel := make(chan Center, 10)
	go displayResult(channel)
	for {
		fmt.Println("Checking availability...")
		date := time.Now().Format("02-01-2006")
		for _, pin := range pincodes {
			wg.Add(1)
			go checkAvailability(pin, date, channel)
		}
		wg.Wait()
		time.Sleep(5 * time.Second)
	}

}

func displayResult(channel chan Center) {
	for c := range channel {
		fmt.Printf("Vaccine available at %v,%v, PIN: %v \n", c.Name, c.BlockName, c.PinCode)
	}
}

func checkAvailability(pin int32, date string, channel chan Center) {

	URL := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/calendarByPin?pincode=%v&date=%v", pin, date)
	//fmt.Println(URL)
	response, _ := http.Get(URL)

	if response.StatusCode == http.StatusOK {
		defer response.Body.Close()
		var centers []Center = decodeData(response)
		for _, c := range centers {
			for _, s := range c.Sessions {
				//Modify here to add more filter conditions like age, vaccine name etc.
				if (doseNo == 1 && s.AvailableCapacityDose1 > 0) || (doseNo == 2 && s.AvailableCapacityDose2 > 0) {
					channel <- c
				}
			}
		}

	}
	wg.Done()
}

func decodeData(response *http.Response) []Center {
	bytes, _ := ioutil.ReadAll(response.Body)
	data := string(bytes)
	decoder := json.NewDecoder(strings.NewReader(data))
	_, err := decoder.Token()
	if err != nil {
		fmt.Println("Error while decoding response.")
	}
	var centers Centers
	json.Unmarshal(bytes, &centers)
	return centers.Centers
}

type Centers struct {
	Centers []Center
}

type Center struct {
	Name      string    `json:"name"`
	BlockName string    `json:"block_name"`
	PinCode   int32     `json:"pincode"`
	Sessions  []Session `json:"sessions"`
}
type Session struct {
	SessionId              string   `json:"session_id"`
	Date                   string   `json:"date"`
	MinAgeLimit            int32    `json:"min_age_limit"`
	Vaccine                string   `json:"vaccine"`
	Slots                  []string `json:"slots"`
	AvailableCapacityDose1 int32    `json:"available_capacity_dose1"`
	AvailableCapacityDose2 int32    `json:"available_capacity_dose2"`
}
