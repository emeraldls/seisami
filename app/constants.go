package main

import "os"

func getCloudApiUrl() string {
	fixedCloudUrl := "https://cloud.seisami.hooklytics.com"

	apiUrl := os.Getenv("CLOUD_API_URL")
	if apiUrl != "" {
		return apiUrl
	}

	return fixedCloudUrl
}

func getWebUrl() string {
	fixedWebUrl := "https://seisami.hooklytics.com"

	webUrlEnv := os.Getenv("WEB_URL")
	if webUrlEnv != "" {
		return webUrlEnv
	}
	return fixedWebUrl

}
