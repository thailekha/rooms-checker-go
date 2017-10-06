package main

import (
	"os"
	"io"
	"log"
	"sort"
	"time"
	"bufio"
	"os/exec"
	"strconv"
	s "strings"
	console "fmt"
	"github.com/PuerkitoBio/goquery"
)

var MAX_CMDS = 5
var cmds = 0

var ROOMS = []string{"IT101","IT102","IT103","IT118","IT119","IT120","IT201","IT202","IT203","IT220","IT221","IT222","ITG01","ITG02","ITG03","ITG17","ITG18","ITG19",}
var MON = []int{2,3,4,5,6,7,8,9,10}
var TUE = []int{11,12,13,14,15,16,17,18,19}
var WED = []int{20,21,22,23,24,25,26,27,28}
var THU = []int{29,30,31,32,33,34,35,36,37}
var FRI = []int{38,39,40,41,42,43,44,45,46}

var DAY_LOOKUP = map[string][]int {
	"1": MON,
	"2": TUE,
	"3": WED,
	"4": THU,
	"5": FRI,
}

var TIME_LOOKUP = map[string]string {
	"1":"09:15",
	"2":"10:15",
	"3":"11:15",
	"4":"12:15",
	"5":"13:15",
	"6":"14:15",
	"7":"15:15",
	"8":"16:15",
}

var TIME = "td:nth-child(1) > small"
var MOD = "td:nth-child(2) > small"

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func clockwise(a string, b string) bool {
	aT, e1 := strconv.ParseInt(s.Split(a,":")[0],10,0)
	check(e1)
	bT, e2 := strconv.ParseInt(s.Split(b,":")[0],10,0)
	check(e2)

	return bT >= aT
}

func lookupTime(key string) []string {
	res := make([]string, 0)
	baseTime := TIME_LOOKUP[key]

	for _, time := range TIME_LOOKUP {
		if clockwise(baseTime, time) {
			res = append(res, time)
		}
	}

	sort.Strings(res)
	return res
}

func contains(time string, times []string) bool {
	for _, t := range times {
		if time == t {
			return true
		}
	}
	return false
}

func getSelector(param int) string {
	return "#divTT > table:nth-child(2) > tbody > tr:nth-child(" + strconv.Itoa(param) + ")"
}

func main() {

	//input
	d, t := prompt()
	day, times := DAY_LOOKUP[d], lookupTime(t)
	console.Print("day: ")
	console.Println(day)
	console.Print("times: ")
	console.Println(times)

	channel := make(chan []string)

	//do query for each room
	for _, room := range ROOMS {
		go process(day, times, room, channel)
	}

	dataCount := 0
	for freeTimes := range channel {
		cmds = cmds - 2
		dataCount = dataCount + 1

		if len(freeTimes) > 1 {
			sort.Strings(freeTimes)
			console.Println(freeTimes)
		}

		if dataCount == len(ROOMS) {
			close(channel)
		}
	}
}

func process(day []int, times []string, room string, channel chan []string)  {
	for cmds >= MAX_CMDS {
		time.Sleep(70 * time.Millisecond)
	}

	cmds = cmds + 2

	doc, err := goquery.NewDocumentFromReader(getHtml(query(room)))
	check(err)

	freeTimes := []string{room}

	for index, param := range day {
		if index > 0 { 
			doc.Find(getSelector(param)).Each(func(i int, s *goquery.Selection) {
				ti, mo := s.Find(TIME).Text(), s.Find(MOD).Text()
				if contains(ti, times) && len(mo) == 2 {
					freeTimes = append(freeTimes, ti)
				}
			})
		}
	}

	clean(room)

	channel <- freeTimes
}

func query(room string) string {
	_, err := exec.Command("./curlroom.sh", []string{room}...).CombinedOutput()
	check(err)

	return room + ".html"
}

func clean(room string) {
	_, err := exec.Command("rm", []string{room + ".html"}...).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
}

func getHtml(path string) io.Reader {
	file, err := os.Open(path)
	check(err)

	return bufio.NewReader(file)
}

func prompt() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	console.Println("Choose weekday:")
	console.Println("1) Monday ")
	console.Println("2) Tuesday ")
	console.Println("3) Wednesday ")
	console.Println("4) Thursdday ")
	console.Println("5) Friday ")
	console.Print("===> ")

	day, errd := reader.ReadString('\n')
	check(errd)

	console.Println("Choose time:")
	console.Println("1) 09:15 ")
	console.Println("2) 10:15 ")
	console.Println("3) 11:15 ")
	console.Println("4) 12:15 ")
	console.Println("5) 13:15 ")
	console.Println("6) 14:15 ")
	console.Println("7) 15:15 ")
	console.Println("8) 16:15 ")
	console.Print("===> ")

	time, errt := reader.ReadString('\n')
	check(errt)
	

	return s.Split(day, "\n")[0], s.Split(time, "\n")[0]
}