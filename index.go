package findfreetimes

import (
	"bufio"
	ers "errors"
	"io"
	"log"
	"os"
	"sort"
	"time"
	// "runtime"
	console "fmt"
	"github.com/PuerkitoBio/goquery"
	"os/exec"
	"strconv"
	s "strings"
)

// PLAN : cater for all exit status code of curl -> make a wrapper around os exec
// profile for which takes most time

var maxCmds = 200
var cmds = 0
var tolerableCurlErrorCodes = map[string]string{"35": "SSL connect error. The SSL handshaking failed.", "55": "Failed sending network data.", "56": "Failure in receiving network data."}

//var NORMAL_ROOMS = []string{"223","224","225","226","227","228","229","230","AG03","AG04","AG07","AG08","AG09","AG10","AG14","AG15","AG16","AG18","AG20","AG21","AG25","AG26","AG27","AG31","AG32","AG33","AG34","AL1","AL2","AL3","AT103","AT104","AT105","AT107","AT108","AT109","AT110","AT111","AT112","AT121","AT126","AT130","B01","B02","B03","B07","B08","B09","B09A","B10","B11","B12","B13","B15","B16","B18","B19","B20","B21","BETL","BL1","BL14","BL2","BL3","BL4","BL9","BW1","C001","C002","C003","C004","C005","C014","C07","C11","C111","C115","C204","C206","C212","C23","C24","C25","C26","C27","C28","C29","C30","C31","C32","C33","C34","C35","C38","C39","C39A","C42","C47","C48","C48A","C51","CL1","CL2","CL3","CL4","D01","D02","D04","D05","D08","D11","D12","D25","E03","E04","E07","E13","E15","E19A","E19B","ETRC1","ETRC2","ETRC3","F01","F02","F03","F04","F06","F07","F09","F20","F23","F26","F27","F28","F28A","F29","F30","FTG10","FTG11","FTG12","FTG13","FTG14","FTG15","FTG18","FTG19","FTG20","FTG22","FTG23","FTG24","FTG25","FTG29","G12","G17","G18","G19","G20","HA 06","HA 07","HA 08","HA 17","HA 18","HA 21","HA 22","TL114","TL116","TL120","TL121","TL128","TL129","TL157","TL158","TL159","TL221","TL225","TL228","TL235","TL236","TL238","TL244(A)","TL244(B)","TL245","TL249","TL250","TL251","TL252","W02","W03","W04","W05","W06","W07","W08","W09","W10","W11","W12","W13","W14","W18","W19","W20","W21",}
var rooms = []string{"C001", "C002", "C003", "C004", "C005", "C014", "C07", "C11", "C111", "C115", "C204", "C206", "C212", "C23", "C24", "C25", "C26", "C27", "C28", "C29", "C30", "C31", "C32", "C33", "C34", "C35", "C38", "C39", "C39A", "C42", "C47", "C48", "C48A", "C51", "CL1", "CL2", "CL3", "CL4", "D01", "D02", "D04", "D05", "D08", "D11", "D12", "D25", "E03", "E04", "E07", "E13", "E15", "E19A", "E19B", "ETRC1", "ETRC2", "ETRC3", "F01", "F02", "F03", "F04", "F06", "F07", "F09", "F20", "F23", "F26", "F27", "F28", "F28A", "F29", "F30", "FTG10", "FTG11", "FTG12", "FTG13", "FTG14", "FTG15", "FTG18", "FTG19", "FTG20", "FTG22", "FTG23", "FTG24", "FTG25", "FTG29", "G12", "G17", "G18", "G19", "G20", "TL114", "TL116", "TL120", "TL121", "TL128", "TL129", "TL157", "TL158", "TL159", "TL221", "TL225", "TL228", "TL235", "TL236", "TL238", "TL244(A)", "TL244(B)", "TL245", "TL249", "TL250", "TL251", "TL252", "W02", "W03", "W04", "W05", "W06", "W07", "W08", "W09", "W10", "W11", "W12", "W13", "W14", "W18", "W19", "W20", "W21"}
var itRooms = []string{"IT101", "IT102", "IT103", "IT118", "IT119", "IT120", "IT201", "IT202", "IT203", "IT220", "IT221", "IT222", "ITG01", "ITG02", "ITG03", "ITG17", "ITG18", "ITG19"}
var supportedTimes = []string{"9:15", "10:15", "11:15", "12:15", "13:15", "14:15", "15:15", "16:15"}
var monRows = []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
var tueRows = []int{11, 12, 13, 14, 15, 16, 17, 18, 19}
var wedRows = []int{20, 21, 22, 23, 24, 25, 26, 27, 28}
var thuRows = []int{29, 30, 31, 32, 33, 34, 35, 36, 37}
var friRows = []int{38, 39, 40, 41, 42, 43, 44, 45, 46}
var timeSelector = "td:nth-child(1) > small"
var moduleSelector = "td:nth-child(2) > small"

func getRows(weekday string) []int {
	switch weekday {
	case "monday":
		return monRows
	case "tuesday":
		return tueRows
	case "wednesday":
		return wedRows
	case "thursday":
		return thuRows
	case "friday":
		return friRows
	default:
		return nil
	}
}

func isTolerableCurlErrorCode(code string) bool {
	//check if the status code is in the list of tolerable Curl Error Codes
	for k, _ := range tolerableCurlErrorCodes {
		if code == k {
			return true
		}
	}
	return false
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func clockwise(a string, b string) bool {
	aT, e1 := strconv.ParseInt(s.Split(a, ":")[0], 10, 0)
	check(e1)
	bT, e2 := strconv.ParseInt(s.Split(b, ":")[0], 10, 0)
	check(e2)

	return bT > aT
}

func validTime(time string) bool {
	for _, t := range supportedTimes {
		if t == time {
			return true
		}
	}
	return false
}

func getTimes(startTime string, endTime string) ([]string, error) {
	if !(validTime(startTime) && validTime(endTime) && (startTime == endTime || clockwise(startTime, endTime))) {
		return nil, ers.New("Invalid time (s)")
	}

	times := []string{startTime}

	for _, t := range supportedTimes {
		if clockwise(startTime, t) {
			append(times, t)
		}

		if t == endTime {
			break
		}
	}

	return times, nil
}

func getSelector(param int) string {
	return "#divTT > table:nth-child(2) > tbody > tr:nth-child(" + strconv.Itoa(param) + ")"
}

func Find(weekday string, startTime string, endTime string, itOnly bool) (string, error) {
	ROOMS := IT_ROOMS

	if itOnly {
		console.Println("IT rooms only")
	} else {
		ROOMS = NORMAL_ROOMS
	}

	channel := make(chan []string)

	//do query for each room
	for _, room := range ROOMS {
		go process(getRows(weekday), getTimes(startTime, endTime), room, channel)
	}

	result := ""

	dataCount := 0
	for freeTimes := range channel {
		cmds = cmds - 1
		dataCount = dataCount + 1

		//console.Println(strconv.Itoa(len(ROOMS) - dataCount) + " rooms left")

		if len(freeTimes) > 1 {
			sort.Strings(freeTimes)
			console.Println(freeTimes)
			result = result + string(freeTimes)
		}

		if dataCount == len(ROOMS) {
			clean()
			close(channel)
		}
	}

	return result
}

func process(day []int, times []string, room string, channel chan []string) {
	for cmds >= maxCmds {
		time.Sleep(70 * time.Millisecond)
	}

	cmds = cmds + 1

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

	channel <- freeTimes
}

func query(room string) string {
	for true {
		// run Curl
		_, err := exec.Command("./curlroom.sh", []string{room}...).CombinedOutput()

		if err != nil {
			errParts := s.Split(string(err.Error()), " ")
			if len(errParts) == 3 && errParts[0] == "exit" && errParts[1] == "status" && isTolerableCurlErrorCode(errParts[2]) {
				//hit either code 55 or 56, the network is probably busy
				console.Println(tolerableCurlErrorCodes[errParts[2]] + ", retrying ...")
				time.Sleep(10 * time.Millisecond)
			} else {
				// neither code 55 nor 56, breaks the program
				log.Fatal(err)
			}
		} else {
			break
		}
	}

	return room + ".html"
}

func clean() {
	_, err := exec.Command("./clean.sh").CombinedOutput()
	check(err)
}

func getHtml(path string) io.Reader {
	file, err := os.Open(path)
	check(err)

	return bufio.NewReader(file)
}

// func nextLine(reader *bufio.Reader) string {
// 	nl, err := reader.ReadString('\n')
// 	check(err)
// 	return s.Split(nl, "\n")[0]
// }
