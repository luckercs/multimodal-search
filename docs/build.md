## (1) server fe build

```shell
cd server-fe 
npm install
npm run build
```

## (2) server be build
```shell
cd server-be
set GOOS=linux
go build -a -o multimodal_search multimodal_search.go
```

## (3) multimodal-search image build
```shell
cd dockerfile
docker build -f multimodal_search.dockerfile -t registry.cn-beijing.aliyuncs.com/luckercs/multimodal-search:1.0 ..
```
