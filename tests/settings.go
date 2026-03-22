package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const timeFormat = "20060102"

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = false
var password = "123456"
var Token = func(p string) string {
	now := time.Now().Format(timeFormat)
	password := sha256.Sum256([]byte(p))
	passwordHash := hex.EncodeToString(password[:])
	tokenData := passwordHash + now
	tokenHash := sha256.Sum256([]byte(tokenData))
	return hex.EncodeToString(tokenHash[:])
}(password)
