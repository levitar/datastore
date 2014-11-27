FROM golang:onbuild

# godoc and vet
RUN go get code.google.com/p/go.tools/cmd/godoc code.google.com/p/go.tools/cmd/vet
