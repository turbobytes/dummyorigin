TAG=$(shell git rev-parse --short HEAD)

all:
	CGO_ENABLED=0 go build -o bin/dummyorigin main.go
	bin/dummyorigin --fetchonly #To make sure local assets directory is populates
ifeq ("$(IMAGE)","")
	$(error IMAGE is not specified)
else
	docker build -t $(IMAGE) .
	docker tag $(IMAGE) $(IMAGE):$(TAG)
endif

clean:
ifeq ("$(IMAGE)","")
	$(error IMAGE is not specified)
else
	docker images $(IMAGE) | grep -Ev 'TAG|latest|none' | awk '{print $2}' | xargs -I '{}' docker rmi '$(IMAGE):{}'
endif
