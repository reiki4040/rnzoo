rnzoo on docker
===

### build rnzoo bin

in above dir
```
bash build_linux.sh
cp rnzoo_linux_amd64 run_on_docker/
```

### build docker image

```
docker build -t rnzoo:0.5.0 .
```

### set env vars

~/.bash_profile or etc...
```
export AWS_ACCESS_KEY_ID="YOUR_KEY"
export AWS_SECRET_ACCESS_KEY="YOUR_SECRET"
export AWS_REGION="ap-northeast-1"
```

### run rnzoo on docker

```
docker run -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} -e AWS_REGION=${AWS_REGION} --rm -it rnzoo:0.5.0 /bin/sh
```
