package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/sirupsen/logrus"
)

/*
	const data = {
	      current_version: "0.4.0",
	      current_version_published: "11 Feb 2024",
	      download_link: "https://github.com/logdyhq/logdy-core/releases/tag/v0.3.0",
	      excerpt: "What's new in version 0.4.0"
	    };
*/

type UpdateResponse struct {
	CurrentVersion          string `json:"current_version"`
	CurrentVersionPublished string `json:"current_version_published"`
	DownloadLink            string `json:"download_link"`
	BlogLink                string `json:"blog_link"`
	Excerpt                 string `json:"excerpt"`
}

func checkUpdates() (error, UpdateResponse) {
	update := UpdateResponse{}
	resp, err := http.Get("https://update.logdy.dev?version=" + Version)
	if err != nil {
		return err, update
	}

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, update
	}

	err = json.Unmarshal(body, &update)

	return err, update
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
	}
}
func checkUpdatesAndPrintInfo() {
	err, update := checkUpdates()

	if err != nil {
		logger.WithField("error", err).Error("Error while checking for Logdy updates")
		return
	}

	if update.CurrentVersion == Version {
		logger.WithFields(logrus.Fields{
			"current_version": Version,
			"latest_version":  update.CurrentVersion,
		}).Debug("No updates detected")
		return
	}

	logger.WithFields(logrus.Fields{
		"response":        update,
		"current_version": Version,
		"latest_version":  update.CurrentVersion,
	}).Debug("New version available")

	logger.Info(Yellow + "----------------------------------------------------------")
	logger.Info(Yellow + ">                NEW LOGDY VERSION AVAILABLE              ")
	logger.Info(Yellow + "> Version: " + update.CurrentVersion)
	logger.Info(Yellow + "> Date published: " + update.CurrentVersionPublished)
	logger.Info(Yellow + "> Download: " + update.DownloadLink)
	logger.Info(Yellow + "> Blog: " + update.DownloadLink)

	if update.Excerpt != "" {
		logger.Info(Yellow + "> " + update.Excerpt)
	}

	logger.Info(Yellow + "----------------------------------------------------------" + Reset)
}
