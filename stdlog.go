package viamstreamdeck

import (
	"log"
	"strings"
	"sync"

	"go.viam.com/rdk/logging"
)

var stdLogOnce sync.Once

// routeStandardLog sends what the streamdeck library writes with Go's standard
// logger through the module's logger at info level. The library announces
// every device it finds and connects to with log.Printf, which goes to
// stderr, and viam-server records everything a module writes to stderr as an
// error. The standard logger is process-wide, so the first resource to be
// built claims it for the whole module.
func routeStandardLog(logger logging.Logger) {
	stdLogOnce.Do(func() {
		log.SetFlags(0)
		log.SetOutput(stdLogWriter{logger: logger.Sublogger("streamdeck-lib")})
	})
}

type stdLogWriter struct {
	logger logging.Logger
}

func (w stdLogWriter) Write(p []byte) (int, error) {
	w.logger.Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
