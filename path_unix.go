// +build !windows

package main

var (
	configDir = "/etc/letsencrypt"
	workDir   = "/var/lib/letsencrypt"
	backupDir = "/var/lib/letsencrypt/backups"
	keyDir    = "/etc/letsencrypt/keys"
	certDir   = "/etc/letsencrypt/certs"
)
