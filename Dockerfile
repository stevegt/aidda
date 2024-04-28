from golang:latest

RUN go install golang.org/x/lint/golint@latest

CMD ["/bin/bash"]
