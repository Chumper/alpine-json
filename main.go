package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	alpineContentsSearchURI = "https://pkgs.alpinelinux.org/packages"
)

type fileInfo struct {
	Package string `json:"package"`
	Version string `json:"version"`
	Branch  string `json:"branch"`
	Repo    string `json:"repository"`
	Arch    string `json:"arch"`
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create a new cli program.
	// Setup the global flags.
	arch := flag.String("arch", "", "Set the arch to use")
	branch := flag.String("branch", "", "Set the branch to use")
	repo := flag.String("repo", "", "Set the repository to search in")
	isVersion := flag.Bool("version", false, "Returns the version of the tool")

	flag.Parse()

	if *isVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		panic(errors.New("must pass a package to search for"))
	}

	query := url.Values{
		"name":   {flag.Args()[0]},
		"branch": {*branch},
		"repo":   {*repo},
		"arch":   {*arch},
	}

	uri := fmt.Sprintf("%s?%s", alpineContentsSearchURI, query.Encode())
	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	jsonResult := []interface{}{}
	files := getFilesInfo(doc)

	for _, f := range files {
		jsonResult = append(jsonResult, f)
	}

	b, err := json.MarshalIndent(jsonResult, "", "  ")
	fmt.Print(string(b))
}

func getFilesInfo(d *goquery.Document) []fileInfo {
	files := []fileInfo{}
	d.Find("#packages > div.table-responsive > table > tbody > tr").Each(func(j int, l *goquery.Selection) {
		f := fileInfo{}
		rows := l.Find("td")
		rows.Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				f.Package = strings.TrimSpace(s.Text())
			case 1:
				f.Version = strings.TrimSpace(s.Text())
			case 4:
				f.Branch = strings.TrimSpace(s.Text())
			case 5:
				f.Repo = strings.TrimSpace(s.Text())
			case 6:
				f.Arch = strings.TrimSpace(s.Text())
			default:
				// logrus.Warnf("Unmapped value for column %d with value %s", i, strings.TrimSpace(s.Text()))
			}
		})
		files = append(files, f)
	})
	return files
}

func in(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
