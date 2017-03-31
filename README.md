# dummyorigin

dummyorigin is a mock origin server to test the behavior of a CDN(or proxy).

The key feature is the arbitrary response header injection based on query string parameters in the request, allowing for great flexibility and ease-of-use. If the server gets request for `/15kb.png?Cache-Control=no-cache&Dummy=Origin` it serves the response with `Cache-Control: no-cache` and `Dummy: Origin` response headers.

dummyorigin comes with support for GZip, range requests, conditional requests, 404 and 301 responses.

## Installation from source

    go get github.com/turbobytes/dummyorigin

## Usage

By default the server loads sample files into assets directory. Use `-nofetch` flag to override this.

    dummyorigin -assets /tmp/dummyassets

Flags

```
-assets string
    Path to asset directory (default "assets")
-fetchonly
    Fetch demo assets and exit, i.e. no server
-http string
    listen addr for http server (default ":80")
-nofetch
    Do not fetch demo assets
```
## Docker

The docker image `turbobytes/dummyorigin` comes pre-populated with sample files

    docker run -it turbobytes/dummyorigin

To build an image with your own assets.

    go get github.com/turbobytes/dummyorigin
    cd $GOPATH/src/github.com/turbobytes/dummyorigin
    mkdir assets
    #At this point, put whatever you want in the assets directory
    IMAGE=username/dummyorigin make all

This gives you a docker image called `username/dummyorigin`

## Server behaviour

When a server gets a request, it does the following.

1. Add response header `Access-Control-Allow-Origin: *`
2. It looks at querystrings and sets headers using key as header name and value as header value. so `?foo=bar` would become `Foo: bar`. Header names are canonicalized as per `net/http`
3. It adds response header `X-Tb-Time` with current time. This is done because most proxies override the `Date` header.
4. [http.ServeFile](https://golang.org/pkg/net/http/#ServeFile). 404 if resource is missing, 301 to `/` for index.html. `http.ServeFile` might decide to overwrite some headers set in step 2
5. [gziphandler.GzipHandler](https://godoc.org/github.com/NYTimes/gziphandler#GzipHandler) from package `github.com/nytimes/gziphandler` handles gzip. 
6. The request and response details are written to console and optionally to [loggly](https://www.loggly.com/) if `LOGGLY_TOKEN` environment variable is set.
