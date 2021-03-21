#Go - Api Load Generator

This is an assessment application I was asked to complete.  I have never written a Go application, so this is my first.  The initial commit took about 6 hours.

I chose Go to do this application because I felt being able to create sub-tasks easily was paramount to performance for this application, and go is very good (and it was easy) to do that.

####Notes:
* The only request that is supported is HTTP GET
* Currently, it just checks for HTTP request status code of 200 as success, anything else is flagged as an error.


#### Additional Work Desired
* Ability to support a terminal dashboard, the screen scrolls so fast the only relevant thing to see is the final line. - HIGH PRIORITY
* Ability to support .ini and cmdline flags simultaneously
* Support other types of requests (POST, PUT, ...)
* Support config file with a list of API commands to send, with expectations


## Installation
With Go installed, the only required is the main.go file in this repository.

## Run
Run un-compiled
`go run ./main.go -host_name=http://127.0.0.1 -host_path=/api/v1/notes -num_workers=2 -num_children=2 -desired_rps=5`


## Command line options
* -host_name={host name} - this is the hostname that is hit for the load testing.  It should be prefaced with http(s)://

* -host_path={host path} - The path that is used for the request.

* -authorization={key:authorization} - This is the authorization information.  It is presumed to be a header key=value combination.

* num_workers={number of workers} - How many workers do we want to spawn off

* num_children={number of children} - How many children per worker do we want to create

* desired_rps={desired responses per second} - This is the target responses per second

## Output
Displays a rolling info of 
* which worker just reported
* the current good/bad count is (good=how many successful requests vs bad=how many failed requests
* Current number of seconds the system has run
* The Desired RPS
* the Current RPS

When complete, it shows
* Total of how many unsuccessful requests
* Total of how many successful requests
* Last RPS calculated
* Desired RPS
* Max RPS attained
