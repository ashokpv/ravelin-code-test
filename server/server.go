package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type websiteData map[string]Data

type Data struct {
	WebsiteURL         string
	SessionID          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool
	FormCompletionTime int
}

type Dimension struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

var (
	Clients = make(map[string]*websiteData)
)

func main() {
	var index bytes.Buffer
	pwd, _ := os.Getwd()
	index.WriteString(pwd)
	index.WriteString("/client/")

	mux := http.NewServeMux()

	// Routes
	mux.Handle("/", http.FileServer(http.Dir(index.String())))
	mux.HandleFunc("/new", NewRequest)
	mux.HandleFunc("/resize", HandlerResize)
	mux.HandleFunc("/submit", FinalSubmit)
	mux.HandleFunc("/copyandpaste", CopyHandler)
	log.Println("Server starting")
	log.Println("Listening on port", ":7080")
	log.Println(http.ListenAndServe(":7080", Middleware(mux)))
}

func Middleware(m http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r)
	})
}

func NewRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Only POST requests are accepted.")
		return
	}
	resp := fetchResponse(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	if _, ok := Clients[url]; !ok {
		c := make(websiteData)
		Clients[url] = &c
	}

	wData, ok := Clients[url]
	if !ok {
		return
	}

	if sessData, ok := (*wData)[sess]; !ok {
		sessData.WebsiteURL = url
		sessData.SessionID = sess

		(*wData)[sess] = sessData
		Clients[url] = wData
		log.Println(sessData)
	}
}

func HandlerResize(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}
	resp := fetchResponse(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)
	resizeFrom, err := fetchDimension(resp["resizeFrom"])
	if err != nil {
		log.Println("Original size is missing")
		return
	}

	resizeTo, err := fetchDimension(resp["resizeTo"])
	if err != nil {
		log.Println("Actual size is missing")
		return
	}

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}
	sessData.ResizeTo = resizeTo
	sessData.ResizeFrom = resizeFrom
	(*wData)[sess] = sessData
	Clients[url] = wData
	log.Println(sessData)
}

func FinalSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("The server only accpets POST requests")
		return
	}

	resp := fetchResponse(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	time := int(resp["time"].(float64))

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}

	sessData.FormCompletionTime = time

	(*wData)[sess] = sessData
	Clients[url] = wData

	log.Println(sessData)
	formattedPrint(sessData)
	hashoutput := DJB(url)

	log.Println("HASHED WEBSITE URL USING DJB ALGORITHM :", hashoutput)
}

func DJB(str string) uint {
	var hash uint = 5381
	for i := 0; i < len(str); i++ {
		hash = ((hash << 5) + hash) + uint(str[i])
	}
	return hash
}

func fetchResponse(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	var response map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("err", err)
	}
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		log.Println("err", err)
	}
	return response
}

func fetchDimension(size interface{}) (dim Dimension, err error) {
	foramtted_json, err := json.Marshal(size)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(foramtted_json, &dim)
	if err != nil {
		log.Println(err)
	}
	return dim, err
}

func formattedPrint(d Data) {
	finalresponse :=
		`
		User Session {{.SessionID}} from {{.WebsiteURL}}
		WebsiteURL: {{.WebsiteURL}}
		ResizeFrom: Width: {{.ResizeFrom.Width}}, Height: {{.ResizeFrom.Height}}
		ResizeTo: Width: {{.ResizeTo.Width}}, Height: {{.ResizeTo.Height}} {{range $key, $value := .CopyAndPaste}}
		CopyAndPaste: FormId: {{$key}}, Paste: {{$value}} {{end}}
		FormCompletionTime: {{.FormCompletionTime}}

`
	tmpl, err := template.New("").Parse(finalresponse)
	if err != nil {
		log.Println(err)
	}
	tmpl.Execute(os.Stdout, d)
}

func CopyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Invalid Request by user. Valid requests are POST.")
		return
	}

	resp := fetchResponse(w, r)
	url := resp["websiteURL"].(string)
	sess := resp["sessionId"].(string)

	formID, ok := resp["formId"].(string) //Id of the field where the copy/paste append
	if !ok {
		log.Println("FormId is needed")
		return
	}

	paste, ok := resp["paste"].(bool) //Boolean: know if event is a paste or not
	if !ok {
		log.Println("Paste field is needed")
		return
	}

	wData, ok := Clients[url]
	if !ok {
		log.Println("Clients[", url, "] not found")
		return
	}

	sessData, ok := (*wData)[sess]
	if !ok {
		log.Println("sessData:", sessData, "not found")
		return
	}

	sessData.CopyAndPaste = make(map[string]bool)
	sessData.CopyAndPaste[formID] = paste

	(*wData)[sess] = sessData
	Clients[url] = wData

	log.Println(sessData)
}
