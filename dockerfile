FROM adelith/work:face-recognition

WORKDIR /go_cv_test

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o go_cv_test /go_cv_test/cmd/main.go

RUN chmod +x /go_cv_test
RUN chmod +r haarcascade_frontalface_default.xml

EXPOSE 8080:8080

CMD [ "./go_cv_test" ]
