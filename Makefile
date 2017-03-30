

all:
	CGO_ENABLED=0 go build -o bin/dummyorigin main.go
	bin/dummyorigin --fetchonly #To make sure local assets directory is populates
ifeq ("$(IMAGE)","")
	$(error IMAGE is not specified)
else
	docker build -t $(IMAGE) .
endif
