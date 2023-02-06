# Check redirects

In a previous job, I had an Nginx with more than thousands of redirects, and we had to verify that each redirect was doing a single jump and that the end path was the correct one, possibly without DOSing the website.

Code is old and it needs some serious refactor.

The only really interesting thing, could be the way I count the redirects, it took me a while to find out how to do it.

## Example

The file with the URIs is a single column with the URIs, for example:
```
$: cat uris.list
/
/redirect/6
/redirect-to?status_code=307&url=http%3A%2F%2Fhttpbingo.org
/relative-redirect/3
/redirect-to?status_code=307&url=http%3A%2F%2Fexample.com%2Fnot-exists
```
The verbose output would be:
```
$: ./check_redirect -base https://httpbingo.org -file uri.list -verbose 
Start processing 5 uris against https://httpbingo.org with host httpbingo.org and 4 workers

"Initial URI","Final URI",Status Code,Nr. Rediretc,Error,Latency
"/","/",200,0,no,329ms
"/redirect-to","/",200,1,no,450.136ms
"/relative-redirect/3","/get",200,3,no,669.214ms
"/redirect-to","/not-exists",404,1,no,426.804ms
"/redirect/6","/get",200,6,no,1.029654s
Total time is 1.030302594s
```

## Full Help
```
Usage of check_redirect
  -base string
    	scheme://host to call
  -file string
    	file to examine, it's a url per line (default "uris.list")
  -host string
    	host header to pass if different from the base
  -http2disable
    	http2 enabled (default true)
  -k	http2 insecure TLS
  -log string
    	log file, if no file is given, the log will be on the stdout
  -num-follow int
    	number of redirects to follow (default 10)
  -num-worker int
    	number of workers (default 4)
  -timeout int
    	timeout in seconds (default 30)
  -user-agent string
    	user agent (default "curl/go.jtheo")
  -verbose
    	verbose
```
