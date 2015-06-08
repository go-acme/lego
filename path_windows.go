// +build !linux

package main

import "os"

var (
	configDir = os.ExpandEnv("${PROGRAMDATA}\\letsencrypt")
	workDir   = os.ExpandEnv("${PROGRAMDATA}\\letsencrypt")
	backupDir = os.ExpandEnv("${PROGRAMDATA}\\letsencrypt\\backups")
	keyDir    = os.ExpandEnv("${PROGRAMDATA}\\letsencrypt\\keys")
	certDir   = os.ExpandEnv("${PROGRAMDATA}\\letsencrypt\\certs")
)
