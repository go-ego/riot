docker 支持
===

## 从源代码 build docker 镜像

1、请从[这里](https://github.com/go-ego/riot/blob/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt?raw=true)下载 weibo_data.txt，放在 testdata/目录下; 可省略

2、进入 examples/codelab 目录

3、建立 docker image

	./build_docker_image.sh 

4、运行 docker container

	docker run -d -p 8080:8080 goriot/riot-codelab

在浏览器中打开 `localhost:8080` 即可打开搜索页面

## 直接从 docker hub 下载镜像

我已经建好了一个 repo 并上传到了 docker hub，用下面的命令 pull 镜像

	docker pull goriot/riot-codelab

下载后运行镜像的方法和上面的相同。
