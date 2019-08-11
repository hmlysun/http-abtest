package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type ZdLogger struct {
	*log.Logger
	dir    string
	format string
	prefix string
}

func NewLogger(dir, format, prefix string) *ZdLogger {
	zdlog := &ZdLogger{
		log.New(os.Stderr, "", log.LstdFlags),
		dir,
		format,
		prefix,
	}

	fileName := zdlog.GetFileName()
	logFile, err := zdlog.CreateLogFile(fileName)
	if err != nil {
		zdlog.Panic(err)
	}

	zdlog.SetOutput(logFile)
	zdlog.SetPrefix(zdlog.prefix)
	zdlog.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	return zdlog
}

func (log *ZdLogger) GetFileName() string {
	return log.dir + time.Now().Format(log.format)
}

func (log *ZdLogger) CreateLogFile(fileName string) (*os.File, error) {
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return logFile, err
}

func (log *ZdLogger) Router(fun func(v ...interface{}), v ...interface{}) {
	go func() {
		fileName := log.GetFileName()
		if !Exists(fileName) {
			logFile, err := log.CreateLogFile(fileName)
			if err != nil {
				log.Panic(err)
			}
			log.SetOutput(logFile)
		}
		fun(v...)
	}()
}

func (log *ZdLogger) RouterFormat(fun func(format string, v ...interface{}), format string, v ...interface{}) {
	go func() {
		fileName := log.GetFileName()
		if !Exists(fileName) {
			logFile, err := log.CreateLogFile(fileName)
			if err != nil {
				log.Panic(err)
			}
			log.SetOutput(logFile)
		}
		fun(format, v...)
	}()
}

func (log *ZdLogger) Print(v ...interface{}) {
	log.Router(log.Logger.Print, v...)
}

func (log *ZdLogger) Printf(format string, v ...interface{}) {
	log.RouterFormat(log.Logger.Printf, format, v...)
}

func (log *ZdLogger) Println(v ...interface{}) {
	log.Router(log.Logger.Println, v...)
}

func (log *ZdLogger) Fatal(v ...interface{}) {
	log.Router(log.Logger.Fatal, v...)
}

func (log *ZdLogger) Fatalf(format string, v ...interface{}) {
	log.RouterFormat(log.Logger.Fatalf, format, v...)
}

func (log *ZdLogger) Fatalln(v ...interface{}) {
	log.Router(log.Logger.Fatalln, v...)
}

func (log *ZdLogger) Panic(v ...interface{}) {
	log.Router(log.Logger.Panic, v...)
}

func (log *ZdLogger) Panicf(format string, v ...interface{}) {
	log.RouterFormat(log.Logger.Panicf, format, v...)
}

func (log *ZdLogger) Panicln(v ...interface{}) {
	log.Router(log.Logger.Panicln, v...)
}
