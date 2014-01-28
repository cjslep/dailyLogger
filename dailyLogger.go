package dailyLogger

import (
	"time"
	"os"
	"io"
	"strings"
	"log"
)

type TimeLogger interface {
	Println(output string)
	Fatal(output string)
}

const (
	LOG_PRINTLN = 0
	LOG_FATAL = 1
)

type basicTimeLogger struct {
	logFile string
	logFilePerms os.FileMode
	logDir string
	currTime time.Time
	quitChan chan bool
	logChan chan logMessage
}

type closableWriter interface {
	io.Writer
	Close() error
}

type logMessage struct {
	MsgType int
	Message []byte
}

func (b *basicTimeLogger) formatTimeString(timeString string) string {
	return strings.Replace(strings.Replace(timeString, " ", "_", -1), ":", "-", -1)
}

func (b *basicTimeLogger) updateIfNewDay() {
	if b.currTime.Day() != time.Now().Local().Day() {
		b.currTime = time.Now().Local()
		b.setNewLogger()
	}
}

func (b *basicTimeLogger) setNewLogger() {
	if (b.quitChan != nil) {
		b.quitChan <- true
		close(b.quitChan)
		close(b.logChan)
	}
	file, err := os.OpenFile(b.logDir + b.formatTimeString(b.currTime.Format(time.Stamp)) + b.logFile + ".txt", os.O_APPEND | os.O_CREATE | os.O_RDWR, b.logFilePerms)
	if err != nil { return }
	b.quitChan = make(chan bool)
	b.logChan = make(chan logMessage)
	go func(f closableWriter, toLog <- chan logMessage, quit <- chan bool) {
		defer f.Close()
		logger := log.New(f, "", log.LstdFlags | log.Lshortfile)
		for {
			select {
				case msg := <- toLog:
					if (msg.MsgType == LOG_PRINTLN) {
						logger.Println(string(msg.Message))
					} else if (msg.MsgType == LOG_FATAL) {
						logger.Fatal(string(msg.Message))
					}
				case <- quit:
					return
			}
		}
	}(file, b.logChan, b.quitChan)
	b.logChan <- logMessage{LOG_PRINTLN, []byte("Logging goroutine successfully launched!")}
}

func (b *basicTimeLogger) logMessage(msgType int, msg string) {
	b.updateIfNewDay()
	b.logChan <- logMessage{msgType, []byte(msg)}
}

func (b *basicTimeLogger) Println(output string) {
	b.logMessage(LOG_PRINTLN, output)
}

func (b *basicTimeLogger) Fatal(output string) {
	b.logMessage(LOG_FATAL, output)
}

func NewBasicTimeLogger(fileLog, dirLog string, filePerms, dirPerms os.FileMode) (t TimeLogger, err error) {
	temp := basicTimeLogger{fileLog, filePerms, dirLog, time.Now().Local(), nil, nil}
	temp.setNewLogger()
	err = os.MkdirAll(dirLog, dirPerms)
	if err != nil { return nil, err }
	return &temp, nil
}