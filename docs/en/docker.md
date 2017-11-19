Docker support
===

## Build docker image from source

1、Please download weibo_data.txt from [here](https://github.com/go-ego/riot/blob/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt?raw=true) and place it in the testdata/directory, Can be omitted.

2、Go to examples/codelab directory

3、Build docker image

	./build_docker_image.sh 

4、Run docker container

	docker run -d -p 8080:8080 unmerged/riot-codelab

Open `localhost: 8080` in your browser to open the search page

## Download the image directly from docker hub

I've built a repo and uploaded it to the docker hub, using the command pull image:

	docker pull unmerged/riot-codelab

The method of running the image after downloading is the same as above.
