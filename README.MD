# Go - Api Load Generator

This is an assessment application I was asked to complete.  I have never written a Go application, so this is my first.  The initial commit took about 12 hours.

I chose Go to do this application because I felt being able to create sub-tasks easily was paramount to performance for this application, and go is very good (and it was easy) to do that.


For the throttling requirement, I used [this - Rate limiting in Golang HTTP client](https://medium.com/mflow/rate-limiting-in-golang-http-client-a22fba15861a) tutorial.


#### Notes:
* Currently, it just checks for HTTP request status code of 200 as success, anything else is flagged as an error and shown on the screen.
* It only sends in the required body as part of the request - i.e. you cannot change the body.  You can change the `name` parameter, but that is all

#### Additional Work Desired
* Ability to support a terminal dashboard, the screen can scroll very fast
* Ability to support .ini and cmdline flags simultaneously
* Support other types of requests (POST, PUT, ...) and other request body's
* Support config file with a list of API commands to send, with expectations



## Installation
Clone this repo

With Go installed, the only required is the main.go file in this repository.

## Run
Run un-compiled

`go run ./main.go -host_name=http://127.0.0.1 -host_path=/api/v1/notes -name=Sven -num_workers=2 -desired_rps=5`

`n {Enter}` while it is running will reset the calculations (for ignoring startup "stuff")

Ctrl-C to stop the program

## Command line options
* -host_name={host name} - this is the hostname that is hit for the load testing.  It should be prefaced with http(s)://

* -host_path={host path} - The path that is used for the request.

* -authorization={key:authorization} - This is the authorization information.  It is presumed to be a header key=value combination.

* num_workers={number of workers} - How many workers do we want to spawn off

* desired_rps={desired responses per second} - This is the target responses per second

* name={name} - Will be used in the body of the POST

## Output
Displays a rolling info of 
* the current good/bad count is (good=how many successful requests vs bad=how many failed requests
* the Average RPS
* The Desired RPS
* The Max RPS that the program has seen
* Current number of seconds the system has run

