package main

import (
	"fmt"
	"os"
	"path/filepath"
	"archive/zip"
	"github.com/alireza0/x-ui/util"
)

func ensureXrayAndGeoFiles() error {
	binDir := "bin"
	os.MkdirAll(binDir, 0755)

	arch := util.GetArch()
	xrayBin := filepath.Join(binDir, "xray")
	geoip := filepath.Join(binDir, "geoip.dat")
	geosite := filepath.Join(binDir, "geosite.dat")

	missing := false
	if _, err := os.Stat(xrayBin); os.IsNotExist(err) {
		missing = true
	}
	if _, err := os.Stat(geoip); os.IsNotExist(err) {
		missing = true
	}
	if _, err := os.Stat(geosite); os.IsNotExist(err) {
		missing = true
	}
	if !missing {
		return nil // All present
	}

	fmt.Println("[x-ui] Downloading xray and geo files for first run...")
	zipUrl := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/v25.9.11/Xray-linux-%s.zip", arch)
	zipPath := filepath.Join(binDir, "Xray-linux-"+arch+".zip")
	if err := util.DownloadFile(zipPath, zipUrl); err != nil {
		return fmt.Errorf("failed to download xray zip: %w", err)
	}
	zipR, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open xray zip: %w", err)
	}
	defer zipR.Close()
	for _, f := range zipR.File {
		outPath := filepath.Join(binDir, f.Name)
		outF, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		inF, err := f.Open()
		if err != nil {
			outF.Close()
			return err
		}
		_, err = io.Copy(outF, inF)
		inF.Close()
		outF.Close()
		if err != nil {
			return err
		}
	}
	os.Remove(zipPath)
	// Download geoip.dat and geosite.dat
	if err := util.DownloadFile(geoip, "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat"); err != nil {
		return fmt.Errorf("failed to download geoip.dat: %w", err)
	}
	if err := util.DownloadFile(geosite, "https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat"); err != nil {
		return fmt.Errorf("failed to download geosite.dat: %w", err)
	}
	return nil
}
