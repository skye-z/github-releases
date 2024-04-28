/*
Using Github to manage versions

BetaX Harbor
Copyright © 2024 SkyeZhang <skai-zhang@hotmail.com>
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// Versioning
type Versioning struct {
	// Github Author
	Author string
	// Github Store
	Store string
	// Product File Name
	Name string
	// Program Restart Command
	Cmd *exec.Cmd
}

type VersionInfo struct {
	Id      int64        `json:"id"`
	Url     string       `json:"html_url"`
	Title   string       `json:"name"`
	Content string       `json:"body"`
	Version string       `json:"tag_name"`
	Date    string       `json:"published_at"`
	Package []AppPackage `json:"assets"`
}

type AppPackage struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Url  string `json:"browser_download_url"`
}

// Get Latest Release Version
func (v Versioning) GetLatestReleaseVersion() *VersionInfo {
	resp, err := http.Get("https://api.github.com/repos/" + v.Author + "/" + v.Name + "/releases/latest")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var info VersionInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil
	}

	return &info
}

// Download New Version
func (v Versioning) DownloadNewVersion() bool {
	info := v.GetLatestReleaseVersion()
	list := info.Package
	name := fmt.Sprintf("harbor_%s_%s", runtime.GOOS, runtime.GOARCH)
	url := ""
	for _, value := range list {
		if value.Name == name {
			url = value.Url
			break
		}
	}
	if url == "" {
		return false
	}

	log.Println("[Update] dwnload update file")
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	out, err := os.Create("update.cache")
	if err != nil {
		return false
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return false
	}
	log.Println("[Update] apply update")
	if err := os.Rename("update.cache", "harbor"); err != nil {
		return false
	}
	return true
}

// Restart With Systemd
func (v Versioning) RestartWithSystemd() error {
	v.Cmd.Stdout = os.Stdout
	v.Cmd.Stderr = os.Stderr

	return v.Cmd.Run()
}
