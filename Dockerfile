FROM scratch

ADD assets /assets

ADD bin/dummyorigin /bin/

ENTRYPOINT ["dummyorigin", "-assets", "/assets"]
