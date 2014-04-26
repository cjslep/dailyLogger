package dailyLogger

import (
	"log"
	"os"
	"strings"
	"time"
	"strconv"
)

const (
	LOG_PRINTLN = iota
	LOG_FATAL
)

type DailyLogger struct {
	currentlyLogging bool // Read only unless in logging goroutine
	logFile      string
	logFilePerms os.FileMode
	logDirPerms os.FileMode
	logDir       string
	currTime     time.Time
	quitChan     chan bool
	logChan      chan logMessage
	newDayChan chan time.Time
}

type logMessage struct {
	MsgType int
	Message []byte
}

func formatTimeString(timeString string) string {
	return strings.Replace(strings.Replace(timeString, " ", "_", -1), ":", "-", -1)
}

func (b *DailyLogger) updateIfNewDay() {
	if b.currTime.Day() != time.Now().Local().Day() {
		b.currTime = time.Now().Local()
		b.newDayChan <- b.currTime
	}
}

func (b *DailyLogger) startLoggingLoop() {
	if b.currentlyLogging {
		return
	}
	if b.quitChan == nil {
		b.quitChan = make(chan bool)
	}
	if b.logChan == nil {
		b.logChan = make(chan logMessage)
	}
	if b.newDayChan == nil {
		b.newDayChan = make(chan time.Time)
	}
	initFile := b.logDir+formatTimeString(b.currTime.Format(time.Stamp))+b.logFile+".log"
	go func(initialFile, logDirectory, logFilename string, dirPerms, filePerms os.FileMode, toLog <-chan logMessage, quit <-chan bool, newTimeChan <-chan time.Time) {
		b.currentlyLogging = true
		err := os.MkdirAll(logDirectory, dirPerms)
		if err != nil {
			b.currentlyLogging = false
			panic("BasicTimeLogger logging goroutine error: " + err.Error())
			return
		}
		file, err := os.OpenFile(initialFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, filePerms)
		if err != nil {
			b.currentlyLogging = false
			panic("BasicTimeLogger logging goroutine error: " + err.Error())
			return
		}
		logger := log.New(file, "", log.LstdFlags)
		for {
			select {
			case msg := <-toLog:
				if msg.MsgType == LOG_PRINTLN {
					logger.Println(string(msg.Message))
				} else if msg.MsgType == LOG_FATAL {
					logger.Fatal(string(msg.Message))
				}
			case newTime := <- newTimeChan:
				file.Close()
				filename := logDirectory+formatTimeString(newTime.Format(time.Stamp))+logFilename+".log"
				file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, filePerms)
				if err != nil {
					b.currentlyLogging = false
					panic("BasicTimeLogger logging goroutine error: " + err.Error())
					return
				}
				logger = log.New(file, "", log.LstdFlags)
				logger.Println("Logging goroutine successfully changed to new file!")
			case <-quit:
				b.currentlyLogging = false
				// Fetch any remaining messages and print them -- may have chosen to quit rather
				// than print any more messages
				for {
					select {
					case msg := <-toLog:
						if msg.MsgType == LOG_PRINTLN {
							logger.Println(string(msg.Message))
						} else if msg.MsgType == LOG_FATAL {
							logger.Fatal(string(msg.Message))
						}
					default:
						break
					}
				}
				file.Close()
				return
			}
		}
	}(initFile, b.logDir, b.logFile, b.logDirPerms, b.logFilePerms, b.logChan, b.quitChan, b.newDayChan)
	b.logChan <- logMessage{LOG_PRINTLN, []byte("Logging goroutine successfully launched!")}
}

func (b *DailyLogger) logMessage(msgType int, msg string) {
	if b.currentlyLogging {
		b.updateIfNewDay()
		b.logChan <- logMessage{msgType, []byte(msg)}
	}
}

func (b *DailyLogger) Println(output string) {
	go b.logMessage(LOG_PRINTLN, output)
}

func (b *DailyLogger) Fatal(output string) {
	go b.logMessage(LOG_FATAL, output)
}

func (b *DailyLogger) Start() {
	if !b.currentlyLogging {
		b.startLoggingLoop()
	}
}

func (b *DailyLogger) Stop() {
	if b.currentlyLogging {
		b.quitChan <- true
	}
}

func NewDailyLogger(fileLog, dirLog string, filePerms, dirPerms os.FileMode) *DailyLogger {
	quoted := strconv.QuoteRuneToASCII(os.PathSeparator)
	pathSeparator := quoted[1:len(quoted)-1] 
	if !strings.HasSuffix(dirLog, pathSeparator) {
		dirLog += pathSeparator
	}
	return &DailyLogger{false, fileLog, filePerms, dirPerms, dirLog, time.Now().Local(), nil, nil, nil}
}

func NewDailyLoggerFromConfig(filename string) (*DailyLogger, error) {
	config, err := LoadLoggingConfig(filename)
	if err != nil {
		return nil, err
	}
	return NewDailyLogger(config.LogFileName, config.DirectoryLogPath, config.FilePermissions, config.FolderPermissions), nil
}