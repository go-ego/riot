docker支持
===

## 从源代码build docker镜像

1、请从[这里](https://github.com/go-ego/gwk/blob/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt?raw=true)下载weibo_data.txt，放在testdata/目录下

2、进入examples/codelab目录

3、建立docker image

	./build_docker_image.sh 

4、运行docker container

	docker run -d -p 8080:8080 unmerged/wukong-codelab

在浏览器中打开 localhost:8080 即可打开搜索页面

## 直接从docker hub下载镜像

我已经建好了一个repo并上传到了docker hub，用下面的命令pull镜像

	docker pull unmerged/wukong-codelab

下载后运行镜像的方法和上面的相同。
