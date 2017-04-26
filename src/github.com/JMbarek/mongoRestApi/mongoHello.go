package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"io"
	"os"
	"net/textproto"
	"image/png"
	"strconv"
)

var (
	session    *mgo.Session
	collection *mgo.Collection
)

type Excursion struct {
	Id                bson.ObjectId  `bson:"_id" json:"excursionnId"`
	OrganizerId       int            `bson:"organizerId" json:"organizerId"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Theme             string         `bson:"theme" json:"theme"`
	State             bool           `json:"state"`
	Adults            int            `json:"adults"`
	Kinder            int            `json:"kinder"`
	Babies            int            `json:"babies"`
	PriceA            float32         `bson:"priceAdult" json:"priceAdult"`
	PriceK            float32         `bson:"priceKind" json:"priceKind"`
	PriceB            float32         `bson:"priceBaby" json:"priceBaby"`
	DepDay            string         `bson:"departureDay" json:"departureDay"`
	DepTime           time.Time      `bson:"departureTime" json:"departureTime"`
	DepPoint          string         `bson:"departurePoint" json:"departurePoint"`
	DepCountry        string         `bson:"departureCountry" json:"departureCountry"`
	Destination       string         `bson:"destination" json:"destination"`
	DestinationRegion string         `bson:"destinationRegion" json:"destinationRegion"`
	Length            string         `json:"length"`
	CreatedOn         time.Time      `bson:"createdOn" json:"createdOn"`
	UpdatedAt         time.Time      `bson:"updatedAt" json:"updatedAt"`
}

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	// contains filtered or unexported fields
}

type ExcursionResource struct {
	Excursion Excursion `json:"excursion"`
}

type ExcursionsResource struct {
	Excursions []Excursion `json:"excursions"`
}

type ThemesResource struct {
	Themes []string `json:"themes"`
}

type DepartureCountriesResource struct {
	DepartureCountries []string `json:"departureCountries"`
}

type RegionsResource struct {
	Regions []string `json:"regions"`
}

type DestinationsInRegionResource struct {
	DestinationsInRegion []string `json:"destinationsInRegion"`
}

func CreateExcursionHandler(w http.ResponseWriter, r *http.Request) {

	var excursionResource ExcursionResource

	err := json.NewDecoder(r.Body).Decode(&excursionResource)
	if err != nil {
		panic(err)
	}

	excursion := excursionResource.Excursion
	// get a new id
	obj_id := bson.NewObjectId()
	excursion.Id = obj_id
	excursion.CreatedOn = time.Now()
	//insert into document collection
	err = collection.Insert(&excursion)
	if err != nil {
		panic(err)
	} else {
		log.Printf("Added new Excursion with title: %s", excursion.Title)
	}
	j, err := json.Marshal(ExcursionResource{Excursion: excursion})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func ExcursionsHandler(w http.ResponseWriter, r *http.Request) {

	var excursions []Excursion

	iter := collection.Find(nil).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func ExcursionByIdHandler(w http.ResponseWriter, r *http.Request) {

	var excursion Excursion
	// Get id from the incoming url
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	err := collection.Find(bson.M{"_id": id}).One(&excursion)
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionResource{Excursion: excursion})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func ThemesHandler(w http.ResponseWriter, r *http.Request) {

	var themes []string

	iter := collection.Find(nil).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		themes = append(themes, result.Theme)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ThemesResource{Themes: themes})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func DepartureCountriesHandler(w http.ResponseWriter, r *http.Request) {

	var departureCountries []string

	iter := collection.Find(nil).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		departureCountries = append(departureCountries, result.DepCountry)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(DepartureCountriesResource{DepartureCountries: departureCountries})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

//TODO
func DestinationsInRegionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	region := vars["region"]

	var destinationsInRegion []string
	iter := collection.Find(bson.M{
		"destinationRegion": region}).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		destinationsInRegion = append(destinationsInRegion, result.Destination)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(DestinationsInRegionResource{DestinationsInRegion: destinationsInRegion})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

//TODO
func RegionsHandler(w http.ResponseWriter, r *http.Request) {

	var regions []string

	iter := collection.Find(nil).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		regions = append(regions, result.DestinationRegion)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(RegionsResource{Regions: regions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func UpdateExcursionHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	// Get id from the incoming url
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	// Decode the incoming note json
	var excursionResource ExcursionResource
	err = json.NewDecoder(r.Body).Decode(&excursionResource)
	if err != nil {
		panic(err)
	}

	// partia update on MogoDB
	err = collection.Update(bson.M{"_id": id},
		bson.M{"$set": bson.M{"title": excursionResource.Excursion.Title,
			"organizerId":         excursionResource.Excursion.OrganizerId,
			"description":         excursionResource.Excursion.Description,
			"theme":               excursionResource.Excursion.Theme,
			"state":               excursionResource.Excursion.State,
			"adults":              excursionResource.Excursion.Adults,
			"kinder":              excursionResource.Excursion.Kinder,
			"babies":              excursionResource.Excursion.Babies,
			"priceAdult":          excursionResource.Excursion.PriceA,
			"priceKind":           excursionResource.Excursion.PriceK,
			"priceBaby":           excursionResource.Excursion.PriceB,
			"departureDay":        excursionResource.Excursion.DepDay,
			"departureTime":       excursionResource.Excursion.DepTime,
			"departurePoint":      excursionResource.Excursion.DepPoint,
			"departureCountry":    excursionResource.Excursion.DepCountry,
			"destination":         excursionResource.Excursion.Destination,
			"destinationRegion":   excursionResource.Excursion.DestinationRegion,
			"length":              excursionResource.Excursion.Length,
			"updatedAt":           time.Now(),
		}})
	if err == nil {
		log.Printf("Updated Excursion: %s", id, excursionResource.Excursion.PriceA)
	} else {
		panic(err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func DeleteExcursionHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]

	// Remove from database
	err = collection.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		log.Printf("Could not find Excursion %s to delete", id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func HandleAPI(w http.ResponseWriter, r *http.Request) {
	// Queries will automatically break down the &variables
	// you don't need to worry about the ampersand & in the
	// URL.
	vars := mux.Vars(r)
	//departureDate := vars["departureDate"]
	departureCountry := vars["departureCountry"]
	destination := vars["destination"]
	theme := vars["theme"]

	var excursions []Excursion
	iter := collection.Find(bson.M{
		//"departureDate":    departureDate,
		"departureCountry": departureCountry,
		"destination":      destination,
		"theme":            theme}).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func HandleAPIByDepartureCountry(w http.ResponseWriter, r *http.Request) {
	// Queries will automatically break down the &variables
	// you don't need to worry about the ampersand & in the
	// URL.
	vars := mux.Vars(r)
	departureCountry := vars["departureCountry"]

	var excursions []Excursion
	iter := collection.Find(bson.M{
		"departureCountry": departureCountry, }).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func HandleAPIByTheme(w http.ResponseWriter, r *http.Request) {
	// Queries will automatically break down the &variables
	// you don't need to worry about the ampersand & in the
	// URL.
	vars := mux.Vars(r)
	theme := vars["theme"]

	var excursions []Excursion
	iter := collection.Find(bson.M{
		"theme": theme}).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func HandleAPIByDestination(w http.ResponseWriter, r *http.Request) {
	// Queries will automatically break down the &variables
	// you don't need to worry about the ampersand & in the
	// URL.
	vars := mux.Vars(r)
	destination := vars["destination"]

	var excursions []Excursion
	iter := collection.Find(bson.M{
		"destination": destination, }).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func HandleAPIByThemes(w http.ResponseWriter, r *http.Request) {

	//vars := mux.Vars(r)
	//theme := vars["theme"]
	// param1 := r.URL.Query().Get("theme")
	param1s := r.URL.Query()["theme"];
	var excursions []Excursion
	iter := collection.Find(bson.M{
		"theme": param1s}).Iter()
	result := Excursion{}
	for iter.Next(&result) {
		excursions = append(excursions, result)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(ExcursionsResource{Excursions: excursions})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	/*if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {*/
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	//}
}

/////////
//ErrDetail ...
type ErrDetail struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
}

//ErrMessage is return message default
type ErrMessage struct {
	Message string      `json:"message"`
	Errors  []ErrDetail `json:"errors"`
}

//SuccessMessage is return Zen message
type SuccessMessage struct {
	Message string `json:"message"`
}
///////77

//HandleUploadImage ...
func HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(2000)
	if err != nil {
		body := ErrMessage{Message: err.Error(), Errors: nil}
		respond(w, r, http.StatusBadRequest, body)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		body := ErrMessage{Message: err.Error(), Errors: nil}
		respond(w, r, http.StatusBadRequest, body)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		body := ErrMessage{Message: err.Error(), Errors: nil}
		respond(w, r, http.StatusBadRequest, body)
		return
	}

	bounds := img.Bounds()
	fmt.Println(bounds.String())
	img = checkSize(img, 30, 30)

	start := time.Now()
	total, results := slidingWindow(img, 30, 30)
	elapsed := time.Since(start)
	log.Printf("slidingWindow took %s", elapsed)
	fmt.Println("resultado:", results)

	data := "File processed with success. File name: " + handler.Filename + " " + bounds.String() + " total sliding=" + strconv.Itoa(total)
	body := SuccessMessage{Message: data}
	respond(w, r, http.StatusOK, body)
}

func main() {
	r := mux.NewRouter()
	//r.Queries("id", "{id:[a-z]+}", "departureCountry", "{departureCountry:[a-z]+}", "destination", "{destination:[a-z]+}", "theme", "{theme:[a-z]+}")
	r.HandleFunc("/api/excursions", ExcursionsHandler).Methods("GET")
	r.HandleFunc("/api/excursions/{id}", ExcursionByIdHandler).Methods("GET")
	r.HandleFunc("/api/excursions", CreateExcursionHandler).Methods("POST")
	r.HandleFunc("/api/excursions/{id}", UpdateExcursionHandler).Methods("PUT")
	r.HandleFunc("/api/excursions/{id}", DeleteExcursionHandler).Methods("DELETE")
	r.HandleFunc("/api/excursions/v1/themes", ThemesHandler).Methods("GET")
	r.HandleFunc("/api/excursions/v1/departureCountries", DepartureCountriesHandler).Methods("GET")
	r.HandleFunc("/api/excursions/v1/regions", RegionsHandler).Methods("GET")
	r.HandleFunc("/api/excursions/v1/destinations", DestinationsInRegionHandler).Queries("region", "{region}").Methods("GET")
	//to handle URL like
	//http://website:8080/api/excursions/v1?departureDate=2017-04-14T00:57:06.625+02:00&departureCountry=Berlinnn&destination=Hamburg&theme=themeeooooooooooooooooooooo
	//r.HandleFunc("/api/excursions", HandleAPI).Queries("departureDate", "{departureDate:[^(19|20)\\d\\d[- /.](0[1-9]|1[012])[- /.](0[1-9]|[12][0-9]|3[01])$]}", "departureCountry", "{departureCountry:[a-z]+}", "destination", "{destination:[a-z]+}", "theme", "{theme:[a-z]+}").Methods("GET")
	//r.HandleFunc("/api/excursions/{version}", HandleAPI).Queries("departureCountry", "{departureCountry:[A-Z][a-z]+}", "destination", "{destination:[A-Z][a-z]+}", "theme", "{theme:[A-Z][a-z]+}").Methods("GET")
	//r.HandleFunc("/api/excursions/{version}", HandleAPI).Queries("departureCountry", "{departureCountry}", "destination", "{destination}", "theme", "{theme}").Methods("GET")
	r.HandleFunc("/api/excursions/v1", HandleAPIByTheme).Queries("theme", "{theme}").Methods("GET")
	r.HandleFunc("/api/excursions/bythemes", HandleAPIByThemes).Methods("GET")
	r.HandleFunc("/api/excursions/v1", HandleAPIByDepartureCountry).Queries("departureCountry", "{departureCountry}").Methods("GET")
	r.HandleFunc("/api/excursions/v1", HandleAPIByDestination).Queries("destination", "{destination}").Methods("GET")
	// upload excursion image
	r.HandleFunc("/upload", HandleUploadImage).Methods("POST")
	//r.HandleFunc("/api/excursions/image/upload", upload)

	http.Handle("/api/", r)
	log.Println("Starting mongodb session")
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	collection = session.DB("excursionsdb").C("excursions")

	log.Println("Listening on 27017")
	http.ListenAndServe(":27017", nil)
}
