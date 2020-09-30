package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

type PageVariables struct {
	Title     string
	Time      string
	Token     string
	CreatedBy string
	URLRoot   string
}

type Record struct {
	AccessCode   string    `json:"AccessCode"`
	CreatedBy    string    `json:"CreatedBy"`
	DateCreated  time.Time `json:"DateCreated"`
	DateRedeemed time.Time `json:"DateRedeemed"`
	DateDisabled time.Time `json:"DateDisabled"`
	IPAddress    string    `json:"IPAddress"`
	Redeemed     bool      `json:"Redeemed"`
	Active       bool      `json:"Active"`
}

func main() {
	http.HandleFunc("/", HomePage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	apiurl := os.Getenv("API_URL")
	apikey := os.Getenv("API_KEY")

	requestdump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestdump))

	client := &http.Client{}
	postdata := make([]byte, 100)
	req, err := http.NewRequest("POST", strings.Join([]string{apiurl, "/generate"}, ""), bytes.NewReader(postdata))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	req.Header.Add("X-Email", r.Header["X-Email"][0])
	req.Header.Add("X-Api-Key", apikey)
	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	fmt.Println(response)

	var record Record
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal([]byte(data), &record)

	if record.AccessCode == "" {
		fmt.Println("Unable to generate token")
	}

	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println(err)
		return
	}

	HomePageVars := PageVariables{
		Title:     "Auth",
		Time:      record.DateCreated.In(pst).Format(time.UnixDate),
		Token:     record.AccessCode,
		CreatedBy: record.CreatedBy,
		URLRoot:   strings.Join([]string{apiurl, "/allowlist"}, ""),
	}

	const html = `
<!DOCTYPE html>
<html>
    <head>
		<title>{{.Title}}</title>
    	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0-rc.2/css/materialize.min.css">
    </head>
    <body>
        <p>Token generated: {{.Token}}</p>
        <p>At: {{.Time}}</p>
        <p>One-time-use URL: {{.URLRoot}}/{{.Token}}</p>
    </body>
</html>`

	t, err := template.New("webpage").Parse(html)
	if err != nil {
		fmt.Println("Got error:")
		fmt.Println(err.Error())
	}

	err = t.Execute(w, HomePageVars)
	if err != nil {
		fmt.Println("Got error:")
		fmt.Println(err.Error())
	}
}
