package overlay

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

//go:embed static
var content embed.FS

type Options struct {
	Url     string
	Port    string
	Timeout int
}

var targetUrl string
var pageTimeout string
var prevDonoAmount float64 = 0.01

func handleIndex(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Error reading embedded HTML file", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/html")

	// Write the HTML content to the response
	_, err = w.Write(htmlContent)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func WebOverlay() {
	http.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, web!")
	})

	http.HandleFunc("/index", handleIndex)

	fmt.Printf("Starting /overlay on port 28080\n")
	err := http.ListenAndServe(":28080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

// Code from original donorbox main.go below

func parseArgs() (*Options, error) {
	options := &Options{}
	flag.StringVar(&options.Url, "url", "https://donorbox.org/support-black-girls-code/fundraiser/christopher-dunaj", "Donorbox URL to check")
	flag.StringVar(&options.Port, "port", "28080", "Port to run the local web server")
	flag.IntVar(&options.Timeout, "timeout", 60, "Page refresh rate, in seconds")
	flag.Usage = func() {
		fmt.Printf("Usage: <app> [options]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	targetUrl = options.Url
	pageTimeout = strconv.Itoa(options.Timeout * 1000)
	return options, nil
}

func DonorboxOverlay() {
	options, err := parseArgs()
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("Server starting on http://localhost:" + options.Port + "\n")
	fmt.Printf("Server checking URL: " + options.Url + "\n")

	http.Handle("/static/images/", http.FileServer(http.FS(content)))

	http.HandleFunc("/donorbox", serveHTML)
	http.ListenAndServe(":"+options.Port, nil)
}

type HTMLVars struct {
	PageTimeout      string
	DonorboxProgress ProgressVars
}

type ProgressVars struct {
	NewDono     bool
	TotalRaised string
	PaidCount   string
	RaiseGoal   string
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	filesys := fs.FS(content)
	tmpl := template.Must(template.ParseFS(filesys, "static/donorbox.html"))
	donorboxProgress, err := getDonorboxProgress()
	if err != nil {
		http.Error(w, "Error getting Donorbox progress", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	result := HTMLVars{
		PageTimeout:      pageTimeout,
		DonorboxProgress: donorboxProgress,
	}
	tmpl.Execute(w, result)
}

func getDonorboxProgress() (ProgressVars, error) {
	// Example code from https://www.makeuseof.com/parse-and-generate-html-in-go/

	//targetUrl := "http://localhost:8080/" // For local testing
	//targetUrl := "https://donorbox.org/support-black-girls-code/fundraiser/christopher-dunaj" // For live testing

	fmt.Println("Fetching URL:", targetUrl)
	resp, err := http.Get(targetUrl)
	if err != nil {
		fmt.Println("Error:", err)
		return ProgressVars{}, err
	}

	defer resp.Body.Close()

	// Use the html package to parse the response body from the request
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return ProgressVars{}, err
	}

	var totalRaised float64
	var paidCount string
	var raiseGoal float64
	var link func(*html.Node)
	link = func(n *html.Node) {

		dollarMatch, _ := regexp.MatchString("^\\$\\d{1,}", n.Data)

		if dollarMatch { //&& n.Type == html.ElementNode {
			for i := range (n.Parent).Attr {
				if (n.Parent).Attr[i].Val == "total-raised" {
					// Formatting the string to remove the dollar sign (https://www.makeuseof.com/go-formatting-numbers-currencies/)
					totalRaised, err = strconv.ParseFloat(n.Data[1:], 64)
					if err != nil {
						fmt.Println("Error:", err)
					}
				}
				if (n.Parent).Attr[i].Val == "bold" {
					raiseGoal, err = strconv.ParseFloat(n.Data[1:], 64)
					if err != nil {
						fmt.Println("Error:", err)
					}
				}
			}
		}

		numMatch, _ := regexp.MatchString("^\\d{1,}", n.Data)

		if n.Data != "" && n.Type == html.TextNode && numMatch {
			for i := range (n.Parent).Attr {
				if (n.Parent).Attr[i].Val == "paid-count" {
					paidCount = n.Data
				}
			}
		}

		// traverses the HTML of the webpage from the first child node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			link(c)
		}
	}

	link(doc)

	fmt.Println("  Number of contributors:", paidCount)
	fmt.Printf("  Total raised: $%g\n", totalRaised)
	fmt.Printf("  Raise goal: $%g\n", raiseGoal)

	prevDonoAmount = totalRaised

	results := ProgressVars{
		NewDono:     prevDonoAmount != 0.01 && prevDonoAmount < totalRaised,
		TotalRaised: fmt.Sprintf("%g", totalRaised),
		PaidCount:   paidCount,
		RaiseGoal:   fmt.Sprintf("%g", raiseGoal),
	}

	return results, nil

}
