package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// getDatedDir builds a directory path logs/{root}/{category}/{success|error}/YYYY/MM/DD
// and ensures it exists.
func getDatedDir(root, category string, success bool, now time.Time) (string, error) {
	resultType := "error"
	if success {
		resultType = "success"
	}

	year := fmt.Sprintf("%04d", now.Year())
	month := fmt.Sprintf("%02d", int(now.Month()))
	day := fmt.Sprintf("%02d", now.Day())

	dir := filepath.Join("logs", root, category, resultType, year, month, day)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// WriteHTTPLogLine writes a single line to the HTTP log file following
// the structure described in docs/LOGGING_STRUCTURE.md:
// logs/app/http/{success|error}/YYYY/MM/DD/http.log
func WriteHTTPLogLine(success bool, line string, now time.Time) {
	dir, err := getDatedDir("app", "http", success, now)
	if err != nil {
		// best-effort only; do not panic on logging failure
		return
	}
	path := filepath.Join(dir, "http.log")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	// Ignore write error; logging is best-effort
	_, _ = f.WriteString(line + "\n")
}

