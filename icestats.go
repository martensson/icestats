package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
	"github.com/oschwald/geoip2-golang"
)

// XML Structs
type listener struct {
	IP        string `xml:"IP"`
	UserAgent string `xml:"UserAgent"`
	Connected string `xml:"Connected"`
	ID        string `xml:"ID"`
	Country   string
	City      string
	Mount     string
}

type source struct {
	XMLNAME   xml.Name   `xml:"source"`
	Mount     string     `xml:"mount,attr"`
	Listener  []listener `xml:"listener"`
	Listeners int        `xml:"Listeners"`
}

type xmlstats struct {
	XMLName xml.Name `xml:"icestats"`
	Source  source   `xml:"source"`
}

// JSON Structs
type stats struct {
	Clients   []listener
	Total     int
	Countries map[string]int
	Cities    map[string]int
}

// Config
type config struct {
	User     string
	Password string
	URL      string
	Mounts   []string
}

var db *geoip2.Reader
var cfg config

func statHandler(w http.ResponseWriter, r *http.Request) {
	stats := new(stats)
	stats.Cities = make(map[string]int)
	stats.Countries = make(map[string]int)
	for _, mount := range cfg.Mounts {
		client := &http.Client{}
		req, err := http.NewRequest("GET", cfg.URL+"/admin/listclients?mount="+mount, nil)
		req.SetBasicAuth(cfg.User, cfg.Password)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			continue
		}
		var xmlstats xmlstats
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		xml.Unmarshal(bodyBytes, &xmlstats)
		stats.Total = stats.Total + xmlstats.Source.Listeners
		for _, l := range xmlstats.Source.Listener {
			ip := net.ParseIP(l.IP)
			record, err := db.City(ip)
			if err != nil {
				continue
			}
			l.City = record.City.Names["en"]
			l.Country = record.Country.Names["en"]
			l.Mount = xmlstats.Source.Mount
			stats.Clients = append(stats.Clients, l)
			if l.Country != "" {
				if i, ok := stats.Countries[l.Country]; ok {
					stats.Countries[l.Country] = i + 1
				} else {
					stats.Countries[l.Country] = 1
				}
			}
			if l.City != "" {
				if i, ok := stats.Cities[l.City]; ok {
					stats.Cities[l.City] = i + 1
				} else {
					stats.Cities[l.City] = 1
				}
			}
		}
	}
	b, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	port := flag.String("p", "8080", "Listening port")
	iface := flag.String("i", "0.0.0.0", "Listening interface")
	configtoml := flag.String("f", "icestats.toml", "Path to config. (default icestats.toml)")
	flag.Parse()
	file, err := ioutil.ReadFile(*configtoml)
	if err != nil {
		log.Fatal(err)
	}
	err = toml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := http.NewServeMux()
	r.HandleFunc("/", statHandler)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	srv := &http.Server{
		Handler: loggedRouter,
		Addr:    *iface + ":" + *port,
		// Good practice: enforce timeouts
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	log.Println("icestats listening on", srv.Addr)
	db, err = geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Fatal(srv.ListenAndServe())
}
