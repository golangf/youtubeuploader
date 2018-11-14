package main

import (
	"fmt"
	"strings"

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

func videoBasics(o *youtube.Video) string {
	var ans = ""
	p := o.Snippet
	s := o.Status
	if p.DefaultLanguage != "" {
		ans += fmt.Sprintf(", &%v", p.DefaultLanguage)
	}
	if p.CategoryId != "" {
		ans += fmt.Sprintf(", #%v", p.CategoryId)
	}
	if s.PrivacyStatus != "" {
		ans += fmt.Sprintf(", :%v", s.PrivacyStatus)
	}
	if s.License != "" {
		ans += fmt.Sprintf(", $%v", s.License)
	} else {
		ans += fmt.Sprintf(", $%v", "standard")
	}
	if s.PublicStatsViewable {
		ans += ", public stats"
	} else {
		ans += ", private stats"
	}
	if s.PublishAt != "" {
		ans += fmt.Sprintf(", T:%v", s.PublishAt)
	}
	if ans != "" {
		ans = ans[2:]
	}
	return ans
}

func videoRecordingDetails(o *youtube.VideoRecordingDetails) string {
	var ans = ""
	if o.LocationDescription != "" {
		ans += fmt.Sprintf(" \"%v\"", o.LocationDescription)
	}
	if o.Location != nil {
		ans += fmt.Sprintf(" (%v, %v)", o.Location.Latitude, o.Location.Longitude)
	}
	if o.RecordingDate != "" {
		ans += fmt.Sprintf(" on %v", o.RecordingDate)
	}
	return ans
}

func logUpload(y *youtube.Video) {
	if !f.Log {
		return
	}
	fmt.Printf("%v\n", y.Snippet.Title)
	fmt.Printf(" - %v\n", videoBasics(y))
	printfString(" @%v\n", videoRecordingDetails(y.RecordingDetails))
	fmt.Printf(" - %v\n", shortString(strings.Join(y.Snippet.Tags, ","), 60))
	printfString("\n%v\n\n", shortString(y.Snippet.Description, 256))
	printfString(" -> id: %v\n", f.ClientID)
	printfString(" -> token: %v\n\n", f.ClientToken)
}
