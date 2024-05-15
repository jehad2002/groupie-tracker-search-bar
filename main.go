package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Data struct {
	A Artist
	R Relation
	L Location
	D Date
}

type Artist struct {
	Id           uint     `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Members      []string `json:"members"`
	CreationDate uint     `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}

type Location struct {
	Locations []string `json:"locations"`
}

type Date struct {
	Dates []string `json:"dates"`
}

type Relation struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

var (
	artistInfo   []Artist
	locationMap  map[string]json.RawMessage
	locationInfo []Location
	datesMap     map[string]json.RawMessage
	datesInfo    []Date
	relationMap  map[string]json.RawMessage
	relationInfo []Relation
)

func ArtistData() []Artist {
	artist, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		log.Fatal(err)
	}
	artistData, err := ioutil.ReadAll(artist.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(artistData, &artistInfo)
	return artistInfo
}

func LocationData() []Location {
	var bytes []byte
	location, err2 := http.Get("https://groupietrackers.herokuapp.com/api/locations")
	if err2 != nil {
		log.Fatal(err2)
	}
	locationData, err3 := ioutil.ReadAll(location.Body)
	if err3 != nil {
		log.Fatal(err3)
	}
	err := json.Unmarshal(locationData, &locationMap)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, m := range locationMap {
		for _, v := range m {
			bytes = append(bytes, v)
		}
	}
	err = json.Unmarshal(bytes, &locationInfo)
	if err != nil {
		fmt.Println("error:", err)
	}
	return locationInfo
}

func DatesData() []Date {
	var bytes []byte
	dates, err2 := http.Get("https://groupietrackers.herokuapp.com/api/dates")
	if err2 != nil {
		log.Fatal(err2)
	}
	datesData, err3 := ioutil.ReadAll(dates.Body)
	if err3 != nil {
		log.Fatal(err3)
	}
	err := json.Unmarshal(datesData, &datesMap)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, m := range datesMap {
		for _, v := range m {
			bytes = append(bytes, v)
		}
	}
	err = json.Unmarshal(bytes, &datesInfo)
	if err != nil {
		fmt.Println("error:", err)
	}
	return datesInfo
}

func RelationData() []Relation {
	var bytes []byte
	relation, err2 := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err2 != nil {
		log.Fatal(err2)
	}
	relationData, err3 := ioutil.ReadAll(relation.Body)
	if err3 != nil {
		log.Fatal(err3)
	}
	err := json.Unmarshal(relationData, &relationMap)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, m := range relationMap {
		for _, v := range m {
			bytes = append(bytes, v)
		}
	}
	err = json.Unmarshal(bytes, &relationInfo)
	if err != nil {
		fmt.Println("error:", err)
	}
	return relationInfo
}

func collectData() []Data {
	ArtistData()
	RelationData()
	LocationData()
	DatesData()
	dataData := make([]Data, len(artistInfo))
	for i := 0; i < len(artistInfo); i++ {
		if artistInfo[i].Id == 21 {
			artistInfo[i].Image = ""
		}
		dataData[i].A = artistInfo[i]
		dataData[i].R = relationInfo[i]
		dataData[i].L = locationInfo[i]
		dataData[i].D = datesInfo[i]
	}
	return dataData
}

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println("Endpoint Hit: returnAllArtists")
	data := ArtistData()
	t, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

func artistPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/artistInfo" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println("Endpoint Hit: Artist's Page")
	value := r.FormValue("ArtistName")
	if value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a := collectData()
	var b Data
	for i, ele := range collectData() {
		if value == ele.A.Name {
			b = a[i]
		}
	}
	t, err := template.ParseFiles("artistPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, b)
}

func returnAllLocations(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("Endpoint Hit: returnAllLocations")
	json.NewEncoder(w).Encode(LocationData())
}

func returnAllDates(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("Endpoint Hit: returnAllDates")
	json.NewEncoder(w).Encode(DatesData())
}

func returnAllRelation(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("Endpoint Hit: returnAllRelation")
	json.NewEncoder(w).Encode(RelationData())
}

// ByName sorts data by artist name.
type ByName []Data

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return strings.Compare(a[i].A.Name, a[j].A.Name) < 0 }

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/search" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}
		results, err := search(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		t, err := template.ParseFiles("searchResults.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, results)
	} else {
		http.NotFound(w, r)
	}
}

func search(query string) ([]Data, error) {
	var results []Data
	query = strings.ToLower(query)
	for _, data := range collectData() {
		if strings.Contains(strings.ToLower(data.A.Name), query) ||
			strings.Contains(strings.ToLower(data.A.FirstAlbum), query) ||
			strings.Contains(fmt.Sprint(data.A.CreationDate), query) {
			if check(data.A.Id, results) {
				results = append(results, data)
			}
			continue
		}
		for _, member := range data.A.Members {
			if strings.Contains(strings.ToLower(member), query) {
				if check(data.A.Id, results) {
					results = append(results, data)
				}
				break
			}
		}
		for _, location := range data.L.Locations {
			if strings.Contains(strings.ToLower(location), query) {
				if check(data.A.Id, results) {
					results = append(results, data)
				}
				break
			}
		}
		for _, date := range data.D.Dates {
			if strings.Contains(strings.ToLower(date), query) {
				if check(data.A.Id, results) {
					results = append(results, data)
				}
				break
			}
		}
		for key := range data.R.DatesLocations {
			concertDates := strings.Join(data.R.DatesLocations[key], " ")
			if strings.Contains(strings.ToLower(concertDates), query) {
				if check(data.A.Id, results) {
					results = append(results, data)
				}
				break
			}
		}
	}
	sort.Sort(ByName(results))

	if len(results) == 0 {
		return nil, errors.New("No results found")
	}
	return results, nil
}

func check(id uint, data []Data) bool {
	for _, m := range data {
		if m.A.Id == id {
			return false
		}
	}
	return true
}

func HandleRequests() {
	fmt.Println("Starting Server at Port 8080")
	fmt.Println("Now open a browser and enter: localhost:8080 into the URL")
	http.HandleFunc("/", homePage)
	http.HandleFunc("/artistInfo", artistPage)
	http.HandleFunc("/locations", returnAllLocations)
	http.HandleFunc("/dates", returnAllDates)
	http.HandleFunc("/relation", returnAllRelation)
	http.HandleFunc("/search", searchHandler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func main() {
	HandleRequests()
}
