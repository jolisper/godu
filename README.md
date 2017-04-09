# godu
du linux command like project in Golang

## Up and Running
```bash
$ go get github.com/jolisper/godu
$ godu /my/dir1/ /my/dir2/
123456 bytes
```
## HTTP Mode
```bash
$ godu --http-mode
$ curl -X POST localhost:8080/size -d '{ "directories": [ "/my/dir1/", "/my/dir2/" ] }'
{"directories":["/my/dir1/", "/my/dir2/"],"totalSize":123456,"unitOfMeasure":"bytes"}
```

