package main

import (
	"log"
	"net"

	"github.com/mmcloughlin/geohash"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "icestats"

var labels = []string{"city", "country", "geohash", "ISO", "Continent", "mount"}

type icestatsCollector struct {
	clientsMetric *prometheus.Desc
}

func newIcestatsCollector() *icestatsCollector {
	return &icestatsCollector{
		clientsMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "clients"),
			"Number of clients connected per location",
			labels, nil,
		),
	}
}

func (collector *icestatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.clientsMetric
}

func (collector *icestatsCollector) Collect(ch chan<- prometheus.Metric) {
	listMounts, err := getStats("/admin/listmounts")
	if err != nil {
		log.Println(err)
		return
	}
	listeners := make(map[string]*clients)
	for _, mount := range listMounts.Source {
		icestats, err := getStats("/admin/listclients?mount=" + mount.Mount)
		if err != nil {
			log.Println(err)
			return
		}
		for _, l := range icestats.Source[0].Listener {
			ip := net.ParseIP(l.IP)
			record, err := db.City(ip)
			if err != nil {
				continue
			}
			cityName := record.City.Names["en"]
			if cityName == "" {
				cityName = "Unknown"
			}
			countryName := record.Country.Names["en"]
			if countryName == "" {
				countryName = "Unknown"
			}
			key := cityName + countryName
			if _, ok := listeners[key]; ok {
				listeners[key].Total = listeners[key].Total + 1
			} else {
				hash := geohash.EncodeWithPrecision(record.Location.Latitude, record.Location.Longitude, 4)
				listeners[key] = &clients{
					ISO:       record.Country.IsoCode,
					Total:     1,
					Continent: record.Continent.Names["en"],
					City:      cityName,
					Country:   record.Country.Names["en"],
					Geohash:   hash,
					Mount:     mount.Mount,
				}
			}
		}
	}
	for _, city := range listeners {
		ch <- prometheus.MustNewConstMetric(collector.clientsMetric, prometheus.GaugeValue, city.Total, city.City, city.Country, city.Geohash, city.ISO, city.Continent, city.Mount)
	}
}
