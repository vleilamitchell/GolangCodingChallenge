package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type Request struct {
	Request string
	Blacklist []string
}

type Response struct {
	Authorized bool
}

func main() {
	HandleRequests()
}

func apiStatic(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.NewEncoder(w).Encode(reqBody)
	fmt.Println("Endpoint Hit: homePage")
}

func processIpRequest(w http.ResponseWriter, r *http.Request){
	reqBody, _ := ioutil.ReadAll(r.Body)
	var request Request
	json.Unmarshal(reqBody, &request)
	log.Println("Request Recieved for IP: " + request.Request)
	var country = getIPCountryName(net.ParseIP(request.Request))
	var response Response
	response.Authorized = !Contains(request.Blacklist, country)
	json.NewEncoder(w).Encode(response)
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", apiStatic)
	router.HandleFunc("/processIpRequest", processIpRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":10000", router))
}

func getIPCountryName(ip net.IP) string{
	db, err := geoip2.Open("data/GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	record, err := db.Country(ip)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("English country name: %v\n", record.Country.Names["en"])
	return record.Country.Names["en"]
}