package dailyLogger

import (
	"testing"
	"strconv"
	"time"
)

func TestConstructor(t *testing.T) {
	temp := NewDailyLogger("", "", 0, 0)
	if temp == nil {
		t.Fail()
	}
}

func TestConstructorBasicInputs(t *testing.T) {
	temp := NewDailyLogger("something", "./", 0666, 0755)
	if temp == nil {
		t.Fail()
	}
}

func TestConstructorStopNoStart(t *testing.T) {
	temp := NewDailyLogger("TestConstructorStopNoStart", "./test", 0666, 0755)
	temp.Stop()
}

func TestConstructorStartStopNoTrailingSeparator(t *testing.T) {
	temp := NewDailyLogger("TestConstructorStartStopNoTrailingSeparator", "./test", 0666, 0755)
	temp.Start()
	temp.Stop()
}

func TestConstructorStartStopTrailingSeparator(t *testing.T) {
	temp := NewDailyLogger("TestConstructorStartStopTrailingSeparator", "./test/", 0666, 0755)
	temp.Start()
	temp.Stop()
}

func TestPrintlnNoStart(t *testing.T) {
	temp := NewDailyLogger("TestPrintlnNoStart", "./test/", 0666, 0755)
	temp.Println("Failed TestPrintlnNoStart if this is in the log.")
}

func TestFatalNoStart(t *testing.T) {
	temp := NewDailyLogger("TestFatalNoStart", "./test/", 0666, 0755)
	temp.Fatal("Failed TestFatalNoStart if this is in the log.")
}

func TestPrintln(t *testing.T) {
	temp := NewDailyLogger("TestPrintln", "./test/", 0666, 0755)
	temp.Start()
	temp.Println("TestPrintln passed if this is in the log.")
	temp.Stop()
}
/*
func TestFatal(t *testing.T) {
	temp := NewDailyLogger("TestFatal", "./test/", 0666, 0755)
	temp.Start()
	temp.Fatal("TestFatal passed if this is in the log.")
	temp.Stop()
}*/

func TestConcurrentPrintlnsNotAllMakeIt(t *testing.T) {
	temp := NewDailyLogger("TestConcurrentPrintlnsNotAllMakeIt", "./test/", 0666, 0755)
	temp.Start()
	for i := 0; i < 500; i++ {
		go temp.Println("TestPrintln passed if this is in the log! Number=" + strconv.Itoa(i))
	}
	temp.Stop()
}

func TestConcurrentPrintlns(t *testing.T) {
	temp := NewDailyLogger("TestConcurrentPrintlns", "./test/", 0666, 0755)
	temp.Start()
	for i := 0; i < 500; i++ {
		go temp.Println("TestPrintln passed if this is in the log! Number=" + strconv.Itoa(i))
	}
	time.Sleep(1 * time.Second)
	temp.Stop()
}