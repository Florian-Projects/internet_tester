package main

import (
	"bufio"
	"fmt"
	"github.com/go-ping/ping"
	"os"
	"time"
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func scheduler(destination string, ch chan *ping.Statistics) {
	for {
		go sendPing(5, destination, ch)
		time.Sleep(30 * time.Second)
	}
}

func sendPing(n int, destination string, ch chan *ping.Statistics) {
	pinger, err := ping.NewPinger(destination)
	if err != nil {
		panic(err)
	}

	pinger.Count = n
	pinger.Timeout = 5 * time.Second
	pinger.Run()
	ch <- pinger.Statistics()
	return
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <destinationFilePath> <resultLogFilePath>")
		return
	}

	destinationsFilePath := os.Args[1]
	resultLogFilePath := os.Args[2]
	lines, err := readLines(destinationsFilePath)
	ch := make(chan *ping.Statistics, 50)
	if err != nil {
		fmt.Println("Error reading file")
	}

	// spawn a scheduler for each destination
	for _, line := range lines {
		go scheduler(line, ch)
	}
	resultFile, err := os.OpenFile(resultLogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer resultFile.Close()

	for result := range ch {
		influxLogLine := fmt.Sprintf("ping_result,dest_name=%s,dest_ip=%s packet_send=%d,packet_recv=%d,packet_loss=%f,avg_rtt%d,min_rtt=%d,max_rtt=%d,std_dev_rtt=%d %d\n", result.Addr, result.IPAddr.String(), result.PacketsSent, result.PacketsRecv, result.PacketLoss, result.AvgRtt.Milliseconds(), result.MinRtt.Milliseconds(), result.MaxRtt.Milliseconds(), result.StdDevRtt.Milliseconds(), time.Now().UnixNano())
		_, err := resultFile.WriteString(influxLogLine)
		if err != nil {
			return
		}
	}
}
