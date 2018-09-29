# Random top links from Hacker News

## Dependencies
go get github.com/gorilla/mux

You don't *really* need gorilla/mux for this use case, but it's easier to add features down the line.

## Usage

Starts a server that displays ten random top Hacker News stories. Run it with **grab** and it'll grab the latest top stories, overwriting any existing JSON files with updated versions.

The server only loads the files when it starts. You'll have to restart it to put new stories in the chaos.

Defaults to port 8000. 

Start the server and load **localhost:8000** in your browser.

The current 1400 or so JSON files are meant to be a demo! All came from the API, and they work, but I don't plan to update them in the GitHub repository.