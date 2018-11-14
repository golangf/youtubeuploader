package main

import (
	"fmt"

	youtube "google.golang.org/api/youtube/v3"
)

func printfString(msg string, val string) {
	if val != "" {
		fmt.Printf(msg, val)
	}
}

func logf(msg string, a ...interface{}) {
	if f.Log {
		fmt.Printf(msg, a...)
	}
}

func logfString(msg string, val string) {
	if f.Log && val != "" {
		fmt.Printf(msg, val)
	}
}

func logFlags() {
	if !f.Log {
		return
	}
	printfString("Video: %v\n", f.Video)
	printfString("Thumbnail: %v\n", f.Thumbnail)
	printfString("Caption: %v\n", f.Caption)
	printfString("Description Path: %v\n", f.DescriptionPath)
	printfString("Meta: %v\n", f.Meta)
	printfString("Client ID: %v\n", f.ClientID)
	printfString("Client Token: %v\n", f.ClientToken)
}

func logUpload(y *youtube.Video) {
	if !f.Log {
		return
	}
	printfString("Title: %v\n", y.Snippet.Title)
	printfString("%v\n", shortString(y.Snippet.Description, 256))
	fmt.Printf("Tags: %v\n", y.Snippet.Tags)
	printfString("Default Language: %v\n", y.Snippet.DefaultLanguage)
	printfString("Default Audio Language: %v\n", y.Snippet.DefaultAudioLanguage)
	printfString("Category ID: %v\n", y.Snippet.CategoryId)
	printfString("Privacy Status: %v\n", y.Status.PrivacyStatus)
	printfString("License: %v\n", y.Status.License)
	fmt.Printf("Public Stats Viewable: %v\n", y.Status.PublicStatsViewable)
	printfString("Publist At: %v\n", y.Status.PublishAt)
	printfString("Recording Date: %v\n", y.RecordingDetails.RecordingDate)
	if y.RecordingDetails.Location != nil {
		fmt.Printf("Location Latitude: %v\n", y.RecordingDetails.Location.Latitude)
		fmt.Printf("Location Longitude: %v\n", y.RecordingDetails.Location.Longitude)
	}
	printfString("Location Description: %v\n", y.RecordingDetails.LocationDescription)
}
