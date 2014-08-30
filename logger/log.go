/*
 * Logging-related functions.
 *
 * (c) 2011-2012 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package logger

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"fmt"
	"os"
	"time"
)

///////////////////////////////////////////////////////////////////////
// Logging constants

const (
	CRITICAL = iota // critical errors
	SEVERE          // severe errors
	ERROR           // errors
	WARN            // warnings
	INFO            // info
	DBG_HIGH        // debug (high prio)
	DBG             // debug (normal)
	DBG_ALL         // debug (all)

	cmd_ROTATE = iota // rotate log file
)

///////////////////////////////////////////////////////////////////////
// Local types

type logger struct {
	msgChan chan string // message to be logged
	cmdChan chan int    // commands to be executed
	logfile *os.File    // current log file (can be stdout/stderr)
	started time.Time   // start time of current log file
	level   int         // current log level
}

///////////////////////////////////////////////////////////////////////
// Local variables

var (
	logInst *logger = nil // singleton logger instance
)

///////////////////////////////////////////////////////////////////////
// Logger-internal methods / functions
/*
 * Instantiate new logger (to stdout) and run its handler loop.
 */
func init() {
	logInst = new(logger)
	logInst.msgChan = make(chan string)
	logInst.cmdChan = make(chan int)
	logInst.logfile = os.Stdout
	logInst.started = time.Now()
	logInst.level = DBG

	go func(){
		for {
			select {
			case msg := <-logInst.msgChan:
				ts := time.Now().Format(time.Stamp)
				logInst.logfile.WriteString(ts + msg)
			case cmd := <-logInst.cmdChan:
				switch cmd {
				case cmd_ROTATE:
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

///////////////////////////////////////////////////////////////////////
// Public logging functions.

/*
 * Punch logging data for given level.
 * @param level int - associated logging level
 * @param line string - information to be logged
 */
func Println(level int, line string) {
	if level <= logInst.level {
		logInst.msgChan <- getTag(level) + line + "\n"
	}
}

//---------------------------------------------------------------------
/*
 * Punch formatted logging data for givel level
 * @param level int - associated logging level
 * @param format string - format definition
 * @param v ...interface{} - list of variables to be formatted
 */
func Printf(level int, format string, v ...interface{}) {
	if level <= logInst.level {
		logInst.msgChan <- getTag(level) + fmt.Sprintf(format, v...)
	}
}

//=====================================================================
// Logfile functions
//=====================================================================

/*
 * Start logging to file.
 * @param filename string - name of logfile
 */
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

//---------------------------------------------------------------------
/*
 * Rotate log file.
 */
func Rotate() {
	logInst.cmdChan <- cmd_ROTATE
}

//=====================================================================
// Human-readable log tags
//=====================================================================

/*
 * Return numeric log level.
 * @return int - current log level
 */
func GetLogLevel() int {
	return logInst.level
}

//---------------------------------------------------------------------
/*
 * Get current loglevel in human-readable form.
 * @return string - symbolic name of loglevel
 */
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
	case DBG_HIGH:
		return "DBG_HIGH"
	case DBG:
		return "DBG"
	case DBG_ALL:
		return "DBG_ALL"
	}
	return "UNKNOWN_LOGLEVEL"
}

//---------------------------------------------------------------------
/*
 * Set logging level from value
 * @param lvl int - new log level
 */
func SetLogLevel(lvl int) {
	if lvl < CRITICAL || lvl > DBG_ALL {
		Printf(WARN, "[logger] Unknown loglevel '%d' requested -- ignored.\n", lvl)
	}
	logInst.level = lvl
}

//---------------------------------------------------------------------
/*
 * Set logging level from symbolic name.
 * @param name string - name of log level
 */
func SetLogLevelFromName(name string) {
	switch name {
	case "ERROR":
		logInst.level = ERROR
	case "WARN":
		logInst.level = WARN
	case "INFO":
		logInst.level = INFO
	case "DBG_HIGH":
		logInst.level = DBG_HIGH
	case "DBG":
		logInst.level = DBG
	case "DBG_ALL":
		logInst.level = DBG_ALL
	default:
		Println(WARN, "[logger] Unknown loglevel '"+name+"' requested.")
	}
}

//---------------------------------------------------------------------
/*
 * Get loglevel tag as prefix for message
 * @param level int - log level
 * @return string - log tag
 */
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
	case DBG_HIGH:
		return "{D2}"
	case DBG:
		return "{D1}"
	case DBG_ALL:
		return "{D0}"
	}
	return "{?}"
}
