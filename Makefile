

run:
	go run main.go

run-detached:
	tmux new -s mywindow &&\
	go run main.go

return:
	tmux a -t mywindow