package main

import "os"

func getCloudApiUrl() string {
	const CLOUD_API_URL = "http://localhost:8080"
	apiUrl := os.Getenv("CLOUD_API_URL")
	if apiUrl == "" {
		return CLOUD_API_URL //hosting url
	}
	return CLOUD_API_URL
}
