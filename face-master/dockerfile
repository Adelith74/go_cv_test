FROM gocv/opencv:latest

WORKDIR /face-master

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Install dlib dependencies
RUN apt-get install libx11-dev
RUN wget http://dlib.net/files/dlib-19.24.tar.bz2 && \
    tar xvf dlib-19.24.tar.bz2 && \
    cd dlib-19.24/ && \
    mkdir build && \
    cd build && \
    cmake .. && \
    cmake --build . --config Release && \
    make install && \
    ldconfig && \
    cd ..

RUN go build -o face-master /face-master/example/simple-face-recognizer/main.go

RUN chmod +x /face-master

CMD [ "./face-master" ]