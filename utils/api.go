package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"mime"

	"github.com/zalando/go-keyring"
)

var apiClient = NewAPIClient(BaseURL)

const (
	BaseURL     = "http://localhost:8090/api/v1"
	serviceName = "ytrss-cli"
)

type Podcast struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AddUrlRequestBody struct {
	PodcastID string `json:"podcast_id"`
	URL       string `json:"url"`
}

type ConvertRequest struct {
	URLs []string `json:"urls"`
}

type JobResponse struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Title   string `json:"title,omitempty"`
	Created string `json:"created,omitempty"`
}

type UsageResponse struct {
	Usage int `json:"usage"`
	Used  int `json:"used"`
}

func GetApiKey() (string, error) {
	return keyring.Get(serviceName, "api_key")
}

func SetApiKey(apiKey string) error {
	return keyring.Set(serviceName, "api_key", apiKey)
}

func ListPodcasts() ([]Podcast, error) {
	var podcasts []Podcast
	err := apiClient.do("GET", "/list-podcasts", nil, &podcasts)
	if err != nil {
		return nil, err
	}
	return podcasts, nil
}

type Item struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func AddUrlToPodcast(podcastID, url string) (Item, error) {
	requestBody := AddUrlRequestBody{
		PodcastID: podcastID,
		URL:       url,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return Item{}, err
	}

	var item Item
	err = apiClient.do("POST", "/podcasts/add-url", bytes.NewBuffer(jsonBody), &item)
	if err != nil {
		return Item{}, err
	}

	return item, nil
}

func CreateJobs(url string) error {
	requestBody := ConvertRequest{
		URLs: []string{url},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	return apiClient.do("POST", "/convert", bytes.NewBuffer(jsonBody), nil)
}

func GetJob(jobId string) (JobResponse, error) {
	var jobResponse JobResponse
	err := apiClient.do("GET", "/poll/jobs/"+jobId, nil, &jobResponse)
	if err != nil {
		return JobResponse{}, err
	}

	return jobResponse, nil
}

func ListJobs() ([]JobResponse, error) {
	var listJobsResponse struct {
		Jobs []JobResponse `json:"jobs"`
	}

	err := apiClient.do("GET", "/poll/jobs", nil, &listJobsResponse)
	if err != nil {
		return nil, err
	}

	return listJobsResponse.Jobs, nil
}

func DownloadFile(jobID string) (string, error) {
	resp, err := apiClient.download("POST", "/download/"+jobID, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	if disposition == "" {
		disposition = "attachment; filename=" + jobID + ".mp3"
	}

	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return "", err
	}
	filename := params["filename"]

	if _, err := os.Stat("downloads"); os.IsNotExist(err) {
		os.Mkdir("downloads", 0755)
	}

	out, err := os.Create("downloads/" + filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return "downloads/" + filename, nil
}

func GetUsage() (*UsageResponse, error) {
	var usageResponse UsageResponse
	err := apiClient.do("GET", "/get-usage", nil, &usageResponse)
	if err != nil {
		return nil, err
	}
	return &usageResponse, nil
}

func PollItem(itemId string) (Item, error) {
	var item Item
	err := apiClient.do("GET", "/poll/item/"+itemId, nil, &item)
	if err != nil {
		return Item{}, err
	}

	return item, nil
}
