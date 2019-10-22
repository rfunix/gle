package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
)

const version = "0.1.1"
const dateLayout = "2006-01-02 15:04:05"
const sleepTime = 1000

func convertDateToUnixTimestamp(d string) (int64, error) {
	t, error := time.Parse(dateLayout, d)
	if error != nil {
		return 0, error
	}

	ut := t.UnixNano() / int64(time.Millisecond)

	return ut, nil
}

func prettyPrintJSON(j interface{}) {
	json, err := json.MarshalIndent(j, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(json))
}

type command struct {
	logName   string
	apiKey    string
	startDate string
	endDate   string
	query     string
	client    http.Client
}

type logsResponse struct {
	Logs []logResponse `json:"logs"`
}

func (lr *logsResponse) getLogByName(n string) (logResponse, error) {
	for _, l := range lr.Logs {
		if l.Name == n {
			return l, nil
		}
	}
	return logResponse{}, fmt.Errorf("log not found with name %v", n)
}

type logResponse struct {
	LogsetsInfo     []logSetInfo `json:"logsets_info"`
	Name            string       `json:"name"`
	UserData        userData     `json:"user_data"`
	Tokens          []string     `json:"tokens"`
	SourceType      string       `json:"source_type"`
	TokenSeed       interface{}  `json:"token_seed"`
	Structures      []string     `json:"structures"`
	ID              string       `json:"id"`
	RetentionPeriod string       `json:"retention_period"`
	Links           []link       `json:"links"`
}

type logSetInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Links []link `json:"links"`
}

type link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type userData struct {
	LeAgentFilename string `json:"le_agent_filename"`
	LeAgentFollow   string `json:"le_agent_follow"`
}

//PostQueryRequest struct
type PostQueryRequest struct {
	Logs []string `json:"logs"`
	Leql leql     `json:"leql"`
}

type leql struct {
	During    during `json:"during"`
	Statement string `json:"statement"`
}

type during struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type postQueryResponse struct {
	Events   []string `json:"events"`
	ID       string   `json:"id"`
	Leql     leql     `json:"leql"`
	Links    []link   `json:"links"`
	Logs     []string `json:"logs"`
	Progress int32    `json:"progress"`
}

type getQueryResponse struct {
	Events []event  `json:"events"`
	Leql   leql     `json:"leql"`
	Links  []link   `json:"links"`
	Logs   []string `json:"logs"`
}

type event struct {
	Labels         []interface{} `json:"labels"`
	Links          []link        `json:"links"`
	LogID          string        `json:"log_id"`
	Message        string        `json:"message"`
	SequenceNumber int64         `json:"sequence_number"`
	Timestamp      int64         `json:"timestamp"`
}

func (cmd *command) fetchLogs() (*logsResponse, error) {
	var l = new(logsResponse)

	req, err := http.NewRequest("GET", "https://rest.logentries.com/management/logs", nil)
	if err != nil {
		return l, err
	}

	req.Header.Add("x-api-key", cmd.apiKey)

	resp, err := cmd.client.Do(req)

	if err != nil {
		return l, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return l, err
	}

	err = json.Unmarshal(body, &l)

	if err != nil {
		return l, err
	}

	return l, err
}

func (cmd *command) postQuery(logID string) (*postQueryResponse, error) {
	pqrr := new(postQueryResponse)

	from, err := convertDateToUnixTimestamp(cmd.startDate)
	if err != nil {
		return pqrr, err
	}

	to, err := convertDateToUnixTimestamp(cmd.endDate)
	if err != nil {
		return pqrr, err
	}

	pqr := NewPostQueryRequest(logID, cmd.query, from, to)

	b, err := json.Marshal(pqr)

	if err != nil {
		return pqrr, err
	}

	req, err := http.NewRequest("POST", "https://rest.logentries.com/query/logs/", bytes.NewBuffer(b))
	req.Header.Add("x-api-key", cmd.apiKey)
	req.Header.Set("Content-type", "application/json")

	if err != nil {
		return pqrr, err
	}

	resp, err := cmd.client.Do(req)

	if err != nil {
		return pqrr, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return pqrr, err

	}

	err = json.Unmarshal(body, &pqrr)

	if err != nil {
		return pqrr, err
	}

	return pqrr, nil
}

func (cmd *command) getLogMessages(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", cmd.apiKey)

	resp, err := cmd.client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil

}

//NewPostQueryRequest create a new struct from this
func NewPostQueryRequest(logID, statement string, from, to int64) *PostQueryRequest {
	pqr := new(PostQueryRequest)
	pqr.Logs = []string{
		logID,
	}
	pqr.Leql = leql{
		During: during{
			From: from,
			To:   to,
		},
		Statement: statement,
	}
	return pqr
}

func main() {
	cmd := command{}

	app := cli.NewApp()

	app.Name = "gle"
	app.Usage = "logentries cli tool"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "log, l",
			Usage:       "Name of log in logentries",
			Destination: &cmd.logName,
			Required:    true,
		},
		cli.StringFlag{
			Name:        "api-key",
			Usage:       "The logentries api-key, its recomend export envvar with name X_API_KEY",
			Destination: &cmd.apiKey,
			Required:    true,
			EnvVar:      "X_API_KEY",
		},
		cli.StringFlag{
			Name:        "start-date",
			Usage:       "The start date period to search log",
			Destination: &cmd.startDate,
			Required:    true,
		},
		cli.StringFlag{
			Name:        "end-date",
			Usage:       "The end date period to search log",
			Destination: &cmd.endDate,
			Required:    true,
		},
		cli.StringFlag{
			Name:        "query",
			Usage:       "the query to search pattern",
			Destination: &cmd.query,
			Required:    true,
		},
	}

	app.Action = func(c *cli.Context) error {
		if err := run(&cmd); err != nil {
			return err
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func (cmd *command) handleLogs(url string) {
	serializeData := func(body io.Reader, s interface{}) (interface{}, error) {
		b, err := ioutil.ReadAll(body)
		if err != nil {
			return s, err

		}

		err = json.Unmarshal(b, &s)

		if err != nil {
			return s, err
		}

		return s, nil
	}

	resp, err := cmd.getLogMessages(url)

	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		grr := new(getQueryResponse)
		d, err := serializeData(resp.Body, grr)
		if err != nil {
			log.Fatal(err)
		}

		for _, m := range d.(*getQueryResponse).Events {
			fmt.Println(m.Message)
		}

		links := d.(*getQueryResponse).Links
		if len(links) > 0 {
			newURL := d.(*getQueryResponse).Links[0].Href
			time.Sleep(sleepTime * time.Millisecond)
			cmd.handleLogs(newURL)
		}
	case http.StatusAccepted:
		pqr := new(postQueryResponse)
		d, err := serializeData(resp.Body, pqr)
		if err != nil {
			log.Fatal(err)
		}
		links := d.(*postQueryResponse).Links
		if len(links) > 0 {
			newURL := d.(*postQueryResponse).Links[0].Href
			time.Sleep(sleepTime * time.Millisecond)
			cmd.handleLogs(newURL)
		}
	default:
		return
	}

}

func run(cmd *command) error {
	lgs, err := cmd.fetchLogs()
	if err != nil {
		return err
	}

	l, err := lgs.getLogByName(cmd.logName)

	if err != nil {
		return err
	}

	pqr, err := cmd.postQuery(l.ID)
	if err != nil {
		return err
	}

	links := pqr.Links
	if len(links) > 0 {
		url := links[0].Href
		cmd.handleLogs(url)
	}

	return nil
}
