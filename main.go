package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type inArgs struct {
	hostName      string
	hostPath      string
	authorization string
	desiredRPS    int
	numWorkers    int
	numChildren   int
}

func main() {
	sFunc := "main()-->"
	//debug := false

	myArgs := getArgs()
	fmt.Println("myArgs", myArgs)

	// create the workers
	numWorkers := myArgs.numWorkers
	numChildren := myArgs.numChildren
	desiredRPS := myArgs.desiredRPS
	results := make(chan map[string]int, numWorkers*numChildren)
	resultTotals := make(map[int]map[string]int)
	maxRPS := 0.0

	for i := 0; i < numWorkers; i++ {
		go worker(i, myArgs, results)
	}

	// log workers output
	totalGood := 0
	totalBad := 0
	currentRPS := 0.0
	startTime := time.Now()
	completedWorkers := 0

	for {
		val, ok := <-results

		if ok == true {
			totalGood, totalBad = getTotals(resultTotals, val, totalGood, totalBad)
			resultTotals[val["workerId"]] = val
			duration := time.Since(startTime)

			currentRPS = float64(totalGood) / duration.Seconds()
			maxRPS = math.Max(maxRPS, currentRPS)
			fmt.Println(sFunc+"val: ", fmt.Sprintf("workerId: %d bad:%d good:%d", val["workerId"], val["bad"], val["good"]),
				"seconds run", fmt.Sprintf("%.2f", duration.Seconds()),
				"  desiredRPS", desiredRPS,
				"currentRPS", fmt.Sprintf("%.2f", currentRPS),
			)

		} else {
			fmt.Println(sFunc + "worker done")

			completedWorkers++

			if completedWorkers == numWorkers {
				fmt.Println(sFunc + "All the workers have completed")
				break
			}
		}
	}

	// report the results
	fmt.Println(sFunc+"Done",
		", totalBad:", totalBad,
		", totalGood:", totalGood,
		", last RPS:", fmt.Sprintf("%.2f", currentRPS),
		"  Desired RPS:", desiredRPS,
		", Max RPS:", fmt.Sprintf("%.2f", maxRPS),
	)

}

func getTotals(resultTotals map[int]map[string]int, lastVal map[string]int, inTotalGood int, inTotalBad int) (int, int) {
	sFunc := "getTotals()-->"
	debug := false

	lastGood := lastVal["good"]
	lastBad := lastVal["bad"]
	thisWorkersGood := resultTotals[lastVal["workerId"]]["good"]
	thisWorkersBad := resultTotals[lastVal["workerId"]]["bad"]

	totalGood := inTotalGood + (lastGood - thisWorkersGood)
	totalBad := inTotalBad + (lastBad - thisWorkersBad)

	// bulk method
	//for x := 0; x < len(resultTotals); x++ {
	//	totalGood += resultTotals[x]["good"]
	//	if debug {
	//		fmt.Println(sFunc+"x", x, "totalGood", totalGood)
	//	}
	//}

	if debug {
		fmt.Println(sFunc+"last", lastVal)
		fmt.Println(sFunc+"thisResult", resultTotals[lastVal["workerId"]])
		fmt.Println(sFunc+"new inTotalGood", inTotalGood,
			"thisWorkersGood", thisWorkersGood,
			"lastGood", lastGood,
			"inTotalBad", inTotalBad,
			"thisWorkersBad", thisWorkersBad,
			"lastBad", lastBad,
			"\n")
	}

	return totalGood, totalBad
}

func worker(workerId int, myArgs inArgs, results chan<- map[string]int) {
	sFunc := "worker()-->"
	debug := false
	good := 0
	bad := 0
	AuthHeaderKey := ""
	AuthHeaderValue := ""

	fullPath := myArgs.hostName + myArgs.hostPath
	numChildren := myArgs.numChildren
	if len(myArgs.authorization) > 0 {
		auths := strings.SplitN(myArgs.authorization, ":", 2)
		AuthHeaderKey, AuthHeaderValue = auths[0], auths[1]
	}

	if debug {
		fmt.Println(sFunc+"workerId", workerId, "numChildren", numChildren)
		fmt.Println(sFunc+"myArgs", myArgs)
		fmt.Println(sFunc+"auths  key", AuthHeaderKey, "value", AuthHeaderValue)
	}

	for x := 0; x < numChildren; x++ {
		client := &http.Client{} // todo: move this outside the loop for speed?
		if debug {
			fmt.Println(sFunc+"workerId", workerId, "child", x, "fullPath", fullPath)
		}

		req, _ := http.NewRequest("GET", fullPath, nil)
		if len(myArgs.authorization) > 0 {
			req.Header.Add(AuthHeaderKey, AuthHeaderValue)
		}
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(sFunc+"child", x, "err", err)
			bad++
			panic(err)
		}
		_ = resp.Body.Close() // todo: move this outside the loop for speed?

		good++
		if debug {
			fmt.Println(sFunc+"worker id", workerId, "child", x, "resp.status", resp.Status)
			scanner := bufio.NewScanner(resp.Body)
			for i := 0; scanner.Scan() && i < 5; i++ {
				fmt.Println(sFunc+"x", x, "i", strconv.Itoa(i), "text", scanner.Text())
			}
		}
		ret := map[string]int{"workerId": workerId, "good": good, "bad": bad}
		results <- ret
	}

	if debug {
		fmt.Println(sFunc + "ready to return")
	}
	time.Sleep(time.Second) // need to wait for a second so the last results <- goes thru
	close(results)
}

func getArgs() inArgs {
	sFunc := "getArgs()-->"
	debug := false

	a := inArgs{}

	// todo:  wow - this is ugly
	hostName := flag.String("host_name", "localhost", "string of the host name")
	hostPath := flag.String("host_path", "/", "the path to use on the host name")
	authorization := flag.String("authorization", "", "Authorization header info. \"{key}:{value}\" ")
	desiredRPS := flag.Int("desired_rps", 100, "Desired rate per second")
	numWorkers := flag.Int("num_workers", 1, "Num workers to spawn")
	numChildren := flag.Int("num_children", 1, "Num children per worker to spawn")

	flag.Parse()

	a.hostPath = *hostPath
	a.hostName = *hostName
	a.authorization = *authorization
	a.desiredRPS = *desiredRPS
	a.numWorkers = *numWorkers
	a.numChildren = *numChildren

	if debug {
		fmt.Println(sFunc+"returning", a)
		fmt.Println(sFunc+"not understood:", flag.Args())
	}
	return a
}