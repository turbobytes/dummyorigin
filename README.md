# dummyorigin
Mock origin to test header behaviour

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
