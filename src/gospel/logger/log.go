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
	"log"
	"os"
	"fmt"
)

///////////////////////////////////////////////////////////////////////
// Logging constants

const (
	ERROR = iota		// errors
	WARN				// warnings
	INFO				// info
	DBG_HIGH			// debug (high prio)
	DBG					// debug (normal)
	DBG_ALL				// debug (all)
)

///////////////////////////////////////////////////////////////////////
// Public variables

var LogLevel = DBG		// current logging verbosity

///////////////////////////////////////////////////////////////////////
// Public logging functions.

/*
 * Punch logging data for given level.
 * @param level int - associated logging level
 * @param line string - information to be logged
 */
func Println (level int, line string) {
	if level <= LogLevel {
		log.Println (getTag(level) + line)
	}
}

//---------------------------------------------------------------------
/*
 * Punch formatted logging data for givel level
 * @param level int - associated logging level
 * @param format string - format definition
 * @param v ...interface{} - list of variables to be formatted
 */
func Printf (level int, format string, v ...interface{}) {
	if level <= LogLevel {
		log.Print (getTag(level) + fmt.Sprintf (format, v...))
	}
}

//=====================================================================
// Logfile functions
//=====================================================================

/*
 * Start logging to file.
 * @param filename string - name of logfile
 */
func LogToFile (filename string) bool {
	Println (INFO, "[log] file-based logging to '" + filename + "'")
	if f,err := os.Create (filename); err == nil {
		log.SetOutput (f)
		return true
	}
	Println (ERROR, "[log] can't enable file-based logging!")
	return false
}

//=====================================================================
// Human-readable log tags
//=====================================================================

/*
 * Get current loglevel in human-readable form.
 * @return string - symbolic name of loglevel
 */
func GetLogLevel() string {
	switch LogLevel {
		case ERROR:		return "ERROR"
		case WARN:		return "WARN"
		case INFO:		return "INFO"
		case DBG_HIGH:	return "DBG_HIGH"
		case DBG:		return "DBG"
		case DBG_ALL:	return "DBG_ALL"
	}
	return "???"
}

//---------------------------------------------------------------------
/*
 * Get loglevel tag as prefix for message
 * @param level int - log level
 * @return string - log tag
 */
func getTag (level int) string {
	switch level {
		case ERROR:		return "{E}"
		case WARN:		return "{W}"
		case INFO:		return "{I}"
		case DBG_HIGH:	return "{D2}"
		case DBG:		return "{D1}"
		case DBG_ALL:	return "{D0}"
	}
	return "{?}"
}
