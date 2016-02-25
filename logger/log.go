package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	// CRITICAL errors
	CRITICAL = iota
	// SEVERE errors
	SEVERE
	// ERROR message
	ERROR
	// WARN for warning messages
	WARN
	// INFO is for informational messages
	INFO
	// DBG for debug messages
	DBG

	// ROTATE log file command
	ROTATE = iota // rotate log file
)

type logger struct {
	msgChan chan string // message to be logged
	cmdChan chan int    // commands to be executed
	logfile *os.File    // current log file (can be stdout/stderr)
	started time.Time   // start time of current log file
	level   int         // current log level
}

var (
	logInst *logger // singleton logger instance
)

// Instantiate new logger (to stdout) and run its handler loop.
func init() {
	logInst = new(logger)
	logInst.msgChan = make(chan string)
	logInst.cmdChan = make(chan int)
	logInst.logfile = os.Stdout
	logInst.started = time.Now()
	logInst.level = DBG

	go func() {
		for {
			select {
			case msg := <-logInst.msgChan:
				ts := time.Now().Format(time.Stamp)
				logInst.logfile.WriteString(ts + msg)
			case cmd := <-logInst.cmdChan:
				switch cmd {
				case ROTATE:
					if logInst.logfile != os.Stdout {
						fname := logInst.logfile.Name()
						logInst.logfile.Close()
						ts := logInst.started.Format(time.RFC3339)
						os.Rename(fname, fname+"."+ts)
						var err error
						if logInst.logfile, err = os.Create(fname); err != nil {
							logInst.logfile = os.Stdout
						}
						logInst.started = time.Now()
					} else {
						Println(WARN, "[log] log rotation for 'stdout' not applicable.")
					}
				}
			}
		}
	}()
}

// Println punches logging data for given level.
func Println(level int, line string) {
	if level <= logInst.level {
		logInst.msgChan <- getTag(level) + line + "\n"
	}
}

// Printf punches formatted logging data for givel level
func Printf(level int, format string, v ...interface{}) {
	if level <= logInst.level {
		logInst.msgChan <- getTag(level) + fmt.Sprintf(format, v...)
	}
}

// LogToFile starts logging messages to file.
func LogToFile(filename string) bool {
	if logInst.logfile == nil {
		logInst.logfile = os.Stdout
	}
	Println(INFO, "[log] file-based logging to '"+filename+"'")
	if f, err := os.Create(filename); err == nil {
		logInst.logfile = f
		logInst.started = time.Now()
		return true
	}
	Println(ERROR, "[log] can't enable file-based logging!")
	return false
}

// Rotate log file.
func Rotate() {
	logInst.cmdChan <- ROTATE
}

// GetLogLevel returns a numeric log level.
func GetLogLevel() int {
	return logInst.level
}

// GetLogLevelName returns the current loglevel in human-readable form.
func GetLogLevelName() string {
	switch logInst.level {
	case CRITICAL:
		return "CRITICAL"
	case SEVERE:
		return "SEVERE"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DBG:
		return "DBG"
	}
	return "UNKNOWN_LOGLEVEL"
}

// SetLogLevel sets the logging level from numeric value
func SetLogLevel(lvl int) {
	if lvl < CRITICAL || lvl > DBG {
		Printf(WARN, "[logger] Unknown loglevel '%d' requested -- ignored.\n", lvl)
	}
	logInst.level = lvl
}

// SetLogLevelFromName sets the logging level from symbolic name.
func SetLogLevelFromName(name string) {
	switch name {
	case "CRITICAL":
		logInst.level = CRITICAL
	case "SEVERE":
		logInst.level = SEVERE
	case "ERROR":
		logInst.level = ERROR
	case "WARN":
		logInst.level = WARN
	case "INFO":
		logInst.level = INFO
	case "DBG":
		logInst.level = DBG
	default:
		Println(WARN, "[logger] Unknown loglevel '"+name+"' requested.")
	}
}

// GetTag returns the loglevel tag as prefix for message
func getTag(level int) string {
	switch level {
	case CRITICAL:
		return "{C}"
	case SEVERE:
		return "{S}"
	case ERROR:
		return "{E}"
	case WARN:
		return "{W}"
	case INFO:
		return "{I}"
	case DBG:
		return "{D}"
	}
	return "{?}"
}
