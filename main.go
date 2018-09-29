package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const topAPI = "https://hacker-news.firebaseio.com/v0/topstories.json"
const itemAPI = "https://hacker-news.firebaseio.com/v0/item/%s.json"

var urlList []string
var titleList []string
var postIDList []int
var byList []string

func main() {
	if 2 == len(os.Args) && os.Args[1] == "grab" {

		// Grabs top posts from HN API and replaces any that exist with an updated copy
		// This would probably be better with a SQLite DB, but it only loads once, and it's fast...
		resp, err := http.Get(topAPI)

		if err != nil {
			log.Fatal(err)
		}

		// s is short for stories
		var s string

		// b is short for buffer
		b, err := ioutil.ReadAll(resp.Body)
		s = string(b)

		// Using the json package might work better
		s = strings.Replace(s, "[", "", 1)
		s = strings.Replace(s, "]", "", 1)

		var topStories = strings.Split(s, ",")

		for i := range topStories {
			resp, err := http.Get(fmt.Sprintf(itemAPI, topStories[i]))
			if err != nil {
				// We're sending hundreds of requests.
				// If something goes wrong, you need to deal with it.
				// Sacrifice the program to avoid a ban from the API.
				log.Fatal(err)
			}
			b, err := ioutil.ReadAll(resp.Body)

			// There are exactly two situations where Get() won't close the body on its own
			// No harm to overdoing it. Ensures we reuse connections.
			resp.Body.Close()
			dir, err := os.Getwd()

			file, err := os.Create(fmt.Sprintf("%s/top-json/%s.json", dir, topStories[i]))

			if err != nil {
				// Shouldn't happen, so let's not make things worse by trying again...
				log.Fatal("Cannot create file", err)
			}
			// defer in a loop is often a bad idea.
			// The API only sends 500 max.
			// In practice, it's closer to 300-400.
			// You shouldn't run out of file handles.

			defer file.Close()

			fmt.Fprintf(file, string(b))

		}
	} else {
		// default: server start
		var files []string

		jsonRoot := "./top-json"
		err := filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(err)
		}
		type HNPost struct {
			URL   string `json:"url,omitempty"`
			Title string `json:"title,omitempty"`
			Type  string `json:"type,omitempty"`
			By    string `json:"by,omitempty"`
			ID    int    `json:"id,omitempty"`

			// Here's _every_ potential item in the JSON, including those above.
			// For reference and possible future use.

			/*ItemID int `json:"id,omitempty"`
			ItemDeleted     bool   `json:"deleted,omitempty"`
			ItemType        string `json:"type,omitempty"`
			ItemBy          string `json:"by,omitempty"`
			ItemTime        int    `json:"time,omitempty"`
			ItemText        string `json:"text,omitempty"`
			ItemDead        bool   `json:"dead,omitempty"`
			ItemParent      int    `json:"parent,omitempty"`
			ItemPoll        int    `json:"poll,omitempty"`
			ItemKids        []int  `json:"kids,omitempty"`
			ItemURL         string `json:"url,omitempty"`
			ItemScore       int    `json:"score,omitempty"`
			ItemTitle       string `json:"title,omitempty"`
			ItemParts       []int  `json:"parts,omitempty"`
			ItemDescendants []int  `json:"descendants,omitempty"`*/

			// I never tried the int arrays, and I don't know the right way to do it.
			// Mainly a reminder that they can contain multiple items.
		}
		var post HNPost
		for _, file := range files {

			jsonFile, err := os.Open(file)

			if err != nil {
				fmt.Println(err)
			}

			b, err := ioutil.ReadAll(jsonFile)
			jsonFile.Close()
			json.Unmarshal(b, &post)

			// Ask/Show posts had some weird thing going on I didn't want to deal with
			// Feel free to fix it and send a pull request. I made this to find interesting articles!
			if post.Type == "story" && !strings.HasPrefix(post.Title, "Ask HN") && !strings.HasPrefix(post.Title, "Show HN") {
				urlList = append(urlList, post.URL)
				titleList = append(titleList, post.Title)
				byList = append(byList, post.By)
				postIDList = append(postIDList, post.ID)
			}

		}

		router := mux.NewRouter()
		router.HandleFunc("/", TopList).Methods("GET")
		log.Fatal(http.ListenAndServe(":8000", router))

		// server end
	}
}

// TopList returns maxRandom articles from the collection of json files
func TopList(w http.ResponseWriter, r *http.Request) {
	var maxRandom = 10

	var randomURLs []string

	var templateLink = `<a href="%s">%s</a> - (<a href=https://news.ycombinator.com/item?id=%d>HN Link</a>, posted by %s)<hr>`
	w.Write(([]byte)(`<html><meta><title>Random Hacker News Top Links</title></meta><body>`))
	rand.Seed(time.Now().UnixNano())

	p := rand.Perm(len(urlList))
	for _, r := range p[:maxRandom] {
		randomURLs = append(randomURLs, fmt.Sprintf(templateLink, urlList[r], titleList[r], postIDList[r], byList[r]))
	}
	for i := range randomURLs {
		w.Write(([]byte)(randomURLs[i]))

	}
	w.Write(([]byte)(`</body></html>`))

}
