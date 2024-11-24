package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"encoding/json"
	"net/http"
	"net/url"

	"strings"
	"unicode"

	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	if len(os.Getenv("key")) == 0 {
		fmt.Println("key not found in .env")
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		fmt.Println("Usage: ./get_climate \"(City / Lat,Long)\"")
		os.Exit(0)
	}

	query := os.Args[1]

	// https://stackoverflow.com/questions/26722450/remove-diacritics-using-go
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	query, _, _ = transform.String(t, query)

	query = url.QueryEscape(query)
	req := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", os.Getenv("key"), query)

	resp, err := http.Get(req)

	if err != nil {
		log.Fatal(err)
	}

	PrintResponse(resp)

}

func PrintResponse(resp *http.Response) {

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var data struct {
		Location struct {
			Name      string
			LocalTime string
			Region    string
			Country   string
		}
		Current struct {
			Temp_c      float64
			Feelslike_c float64
			Condition   struct {
				Text string
				// a API da opção de identificar condição do clima por código (e.g 1000 -> Sunny), porém não vi a necessidade de mapear os código tudo pra um programa dessa escala e complexidade (pequeno e simples), então meio que resumi a uns 6 casos
			}
		}
	}

	err = json.Unmarshal([]byte(responseBody), &data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("📍 %s\n", data.Location.Name)
	fmt.Printf("🌎 %s\n", data.Location.Country)
	fmt.Printf("🏙️  %s\n", data.Location.Region)
	fmt.Printf("🕒 %s\n", data.Location.LocalTime)
	fmt.Printf("🌡️  %2.2fCº\n", data.Current.Temp_c)

	if data.Current.Feelslike_c > 21 {
		fmt.Printf("🥵 Feels like %2.2fCº!\n", data.Current.Feelslike_c)
	} else {
		fmt.Printf("🥶 Feels like %2.2fCº!\n", data.Current.Feelslike_c)
	}

	condition := cases.Title(language.English).String(data.Current.Condition.Text)

	switch {

	case strings.Contains(condition, "Cloudy") || strings.Contains(condition, "Overcast"):
		fmt.Printf("🌥️  %s\n", condition)
	case strings.Contains(condition, "Rain") || strings.Contains(condition, "Drizzle"):
		fmt.Printf("🌧️  %s\n", condition)
	case strings.Contains(condition, "Snow") || strings.Contains(condition, "Ice") || strings.Contains(condition, "Sleet"):
		fmt.Printf("🌨️  %s\n", condition)
	case strings.Contains(condition, "Thunder"):
		fmt.Printf("⛈️  %s\n", condition)
	case strings.Contains(condition, "Clear") || strings.Contains(condition, "Sunny"):
		fmt.Printf("🌞 %s\n", condition)
	case strings.Contains(condition, "Fog") || strings.Contains(condition, "Mist"):
		fmt.Printf("🌫️  %s\n", condition)
	default:
		fmt.Println("❌ Weather Condition Not found!")
		fmt.Println(condition)
	}

}
