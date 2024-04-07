from golang

RUN go install golang.org/x/lint/golint@latest

CMD ["/bin/bash"]
