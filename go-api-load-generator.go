package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/time/rate"
	"math"
	"net/http"
	"os"
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
	nameToSend    string
	ratePerSecond int
}

func main() {
	sFunc := "main()-->"
	debug := false

	myArgs := getArgs()
	fmt.Printf("myArgs [%+v]\n", myArgs)

	input := make(chan rune, 1)

	// create the workers
	numWorkers := myArgs.numWorkers
	results := make(chan map[string]int)
	resultTotals := make(map[int]map[string]int)
	desiredRPS := myArgs.desiredRPS
	myArgs.ratePerSecond = desiredRPS //int(math.Min(1, float64(desiredRPS/numWorkers)))
	fmt.Println(sFunc+"rps", myArgs.ratePerSecond,
		"desiredRPS", desiredRPS,
		"numWorkers", numWorkers,
	)

	for i := 0; i < numWorkers; i++ {
		go worker(i, myArgs, results)
		//time.Sleep( time.Minute / time.Duration(numWorkers ) )
	}

	// log workers output
	startTime := time.Now()

	completedWorkers := 0
	totalGood := 0
	totalBad := 0
	currentRPS := 0.0
	maxRPS := 0.0

	x := 0
	for {
		val, ok := <-results

		if ok != true {
			fmt.Println(sFunc + "worker done")

			completedWorkers++

			if completedWorkers == numWorkers {
				fmt.Println(sFunc + "All the workers have completed")
				break
			}

			continue
		}
		totalGood, totalBad = getTotals(resultTotals, val, totalGood, totalBad)
		resultTotals[val["workerId"]] = val
		duration := time.Since(startTime)

		currentRPS = float64(totalGood+totalBad) / duration.Seconds()
		maxRPS = math.Max(maxRPS, currentRPS)
		if debug {
			fmt.Println(sFunc+"val: ", fmt.Sprintf("workerId: %4d bad:%4d good:%4d", val["workerId"], val["bad"], val["good"]),
				"seconds run", fmt.Sprintf("%.2f", duration.Seconds()),
				"  desiredRPS", desiredRPS,
				"currentRPS", fmt.Sprintf("%.2f", currentRPS),
			)
		}
		showResults(x, totalBad, totalGood, currentRPS, desiredRPS, maxRPS, duration)

		x++
		go readKey(input)

		select {
		case i := <-input:
			fmt.Println(" got an input", i)
			//os.Exit(1)
			if i == 110 {
				fmt.Println("resetting max now that things are running\n\n\n\n\n")
				maxRPS = 0
				totalBad = 0
				totalGood = 0
				startTime = time.Now().Add(-1 * time.Second)
			}
		case <-time.After(1 * time.Millisecond):
			//fmt.Println("time out")
		}

	}
}

func showResults(num int, totalBad int, totalGood int, currentRPS float64, desiredRPS int, maxRPS float64, duration time.Duration) {
	fmt.Printf(
		">%d- totalBad: [%d], totalGood: [%d], Avg RPS: [%.2f], Desired RPS: [%d], Max RPS: [%.2f], Duration: %f\n",
		num, totalBad, totalGood, currentRPS, desiredRPS, maxRPS, duration.Seconds(),
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
		)
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
	auths := strings.SplitN(myArgs.authorization, ":", 2)
	fmt.Println("auths", auths)
	if auths[0] == myArgs.authorization {
		fmt.Printf("Authorization key:value is not correct   authentication[%s]\n", myArgs.authorization)
		flag.PrintDefaults()
		os.Exit(1)

	} else {
		AuthHeaderKey, AuthHeaderValue = auths[0], auths[1]
	}

	//if debug
	{
		fmt.Println(sFunc+"workerId", workerId,
			"auths  key", AuthHeaderKey,
			"value", AuthHeaderValue,
			"fullPath", fullPath,
			"myArgs", fmt.Sprintf("%+v", myArgs),
		)
	}

	r1 := rate.NewLimiter(
		rate.Every((time.Second/time.Duration(myArgs.ratePerSecond))*time.Duration(myArgs.numWorkers)),
		1,
	)
	fmt.Printf("r1= %+v\n", r1)
	client := NewClient(r1)

	// this works jsonStr := []byte( `{"name": "Mike", "date": "2021-03-22T00:00:00", "requests_sent": "1"}`)

	x := 0
	for {

		values := map[string]string{"name": myArgs.nameToSend,
			"date":          time.Now().Format(time.RFC3339),
			"requests_sent": strconv.Itoa(x)}
		jsonStr, _ := json.Marshal(values)

		req, err := http.NewRequest("POST",
			fullPath,
			bytes.NewBuffer(jsonStr),
		)

		if err != nil {
			fmt.Println("err", err)
		}

		req.Header.Add(AuthHeaderKey, AuthHeaderValue)

		resp, err := client.Do(req, workerId)

		if err != nil {
			fmt.Println(sFunc+"try", x, "err", err)
			bad++
		} else if resp.StatusCode != http.StatusOK {
			fmt.Println(sFunc+"try", x, "http.statusCode", resp.StatusCode) //"resp.Body", resp.Body,
			//"resp.Header", resp.Header,

		} else {
			good++
			if debug {
				fmt.Println(sFunc+"worker id", workerId, "try", x, "resp.status", resp.Status)
				scanner := bufio.NewScanner(resp.Body)
				for i := 0; scanner.Scan() && i < 5; i++ {
					fmt.Println(sFunc+"x", x, "i", strconv.Itoa(i), "text", scanner.Text())
				}
			}

		}

		ret := map[string]int{"workerId": workerId, "good": good, "bad": bad}
		results <- ret
		x++

		if resp != nil {
			_ = resp.Body.Close()
		}
	}

}

func NewClient(r1 *rate.Limiter) *RLHTTPClient {
	c := &RLHTTPClient{
		client:      http.DefaultClient,
		Ratelimiter: r1,
	}

	return c
}

func (c *RLHTTPClient) Do(req *http.Request, workerId int) (*http.Response, error) {
	sFunc := "Do()-->"
	debug := false
	doRateLimiting := true

	//start := time.Now()
	if doRateLimiting {
		ctx := context.Background()
		err := c.Ratelimiter.Wait(ctx) // this is a blocking call. Honors the rate limit
		if err != nil {
			return nil, err
		}
	}
	//fmt.Printf("%d duration [%f]\n", workerId, time.Since(start).Seconds())
	if debug {
		fmt.Println(sFunc+"req", fmt.Sprintf("%+v", req))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getArgs() inArgs {
	sFunc := "getArgs()-->"
	debug := false

	a := inArgs{}

	hostName := flag.String("host_name", "", "string of the host name")
	hostPath := flag.String("host_path", "", "the path to use on the host name")
	authorization := flag.String("authorization", "", "Authorization header info. \"{key}:{value}\" ")
	desiredRPS := flag.Int("desired_rps", 5, "Desired rate per second")
	numWorkers := flag.Int("num_workers", 5, "Num workers to spawn")
	nameToSend := flag.String("name", "Sven", "Name to send in")

	flag.Parse()

	a.hostPath = *hostPath
	a.hostName = *hostName
	a.authorization = *authorization
	a.desiredRPS = *desiredRPS
	a.numWorkers = *numWorkers
	a.nameToSend = *nameToSend

	if debug {
		fmt.Println(sFunc+"returning", fmt.Sprintf("%+v", a))
		fmt.Println(sFunc+"not understood:", flag.Args())
	}

	//validation
	bFault := true
	switch {
	case len(*hostName) == 0:
		fmt.Println("-host_name is required")
	case len(*hostPath) == 0:
		fmt.Println("-host_path is required")
	case len(*authorization) == 0:
		fmt.Println("-authorization is required")

	default:
		bFault = false
	}
	if bFault {
		flag.PrintDefaults()
		os.Exit(1)
	}

	return a
}

type RLHTTPClient struct {
	client      *http.Client
	Ratelimiter *rate.Limiter
}

var reader = bufio.NewReader(os.Stdin)

func readKey(input chan rune) {
	char, _, err := reader.ReadRune()
	if err != nil {
		fmt.Println("Fatal error:", err)
		os.Exit(1)
		//log.Fatal(err)
	}

	input <- char
}
