package logger

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

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
	// FLUSH log (make sure all messages are processed)
	FLUSH
)

type logger struct {
	msgChan chan string // message to be logged
	cmdChan chan int    // commands to be executed
	logfile *os.File    // current log file (can be stdout/stderr)
	started time.Time   // start time of current log file
	level   int         // current log level
	lastMsg string      // last log message
	repeats int         // number of repeats of last message
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
	logInst.lastMsg = ""
	logInst.repeats = 0

	go func() {
		for {
			select {
			case msg := <-logInst.msgChan:
				if msg == logInst.lastMsg {
					logInst.repeats++
					continue
				}
				ts := time.Now().Format(time.Stamp)
				if logInst.repeats > 0 {
					s := fmt.Sprintf("...(last message repeated %d times)\n", logInst.repeats)
					logInst.logfile.WriteString(ts + s)
				}
				logInst.repeats = 0
				logInst.lastMsg = msg
				logInst.logfile.WriteString(ts + msg)
			case cmd := <-logInst.cmdChan:
				switch cmd {
				case ROTATE:
					// Rotate log files
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
				case FLUSH:
					// Flush log messages: This is a functional NOP, as log
					// messages have been processed before this command is
					// handled.
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

// Flush log: make sure all messages are processed.
func Flush() {
	logInst.cmdChan <- FLUSH
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
