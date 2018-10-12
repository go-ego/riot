FROM busybox
EXPOSE 8080
ADD / /
CMD ./search_server \
 --weibo_data=weibo_data.txt \
 --dict_file=dictionary.txt \
 --stop_token_file=stop_tokens.txt \
 --static_folder=static
