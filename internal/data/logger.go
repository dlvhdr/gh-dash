package data

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/log/v2"
)

type HTTPLogger struct {
	strBuilder *strings.Builder
	limit      int
}

// NewHTTPLogger creates a new HTTPLogger.
// If limit is 0 the limit is 100000.
func NewHTTPLogger(limit int) HTTPLogger {
	l := limit
	if l == 0 {
		limit = 100000
	}
	logger := HTTPLogger{
		strBuilder: &strings.Builder{},
		limit:      limit,
	}
	return logger
}

func (logger *HTTPLogger) Write(p []byte) (n int, err error) {
	s := string(p)
	if strings.HasPrefix(s, "* Request at") {
		logger.strBuilder.Reset()
	}

	lp := len(p)
	n = min(lp, logger.limit)
	n, err = logger.strBuilder.Write(p[:n])
	if lp > logger.limit {
		logger.strBuilder.WriteString("\n------------------ TRIMMED ------------------\n")
		fmt.Fprintf(logger.strBuilder, "Content length: %d (limit is %d)\n", lp, logger.limit)
		return logger.limit, errors.New("msg too long")
	}

	if strings.HasPrefix(s, "* Request took") {
		log.Debug("HTTPLogger", "msg", logger.strBuilder.String())
	}

	return n, err
}
