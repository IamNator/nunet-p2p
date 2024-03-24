

run:
	go run main.go

run-detached:
	tmux new -s mywindow &&\
	go run main.go

return:
	tmux a -t mywindow

fmt:
	goimports-reviser -rm-unused -use-cache -set-alias -format ./...