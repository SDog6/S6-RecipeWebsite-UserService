
#base image to be used
FROM golang:1.21.0-alpine as builder
#select app folder
WORKDIR /S6-RecipeWebsite-UserService
#copy go file to install packages
#install dependencies
COPY go.* ./
RUN go mod download
# Copy local code to the container image.
COPY . ./

#copy source code to image, ignore node_modules because we already installed them
#COPY . /react-frontend

# Build the binary.
RUN go build -o main .
#set environment
ENV port=3000
#expose port so we can access the app
EXPOSE 3000
# Add Datadog label for log collection
LABEL "com.datadoghq.ad.logs"='[{"source": "goapp", "service": "user-service"}]'
#command to start the app , "PORT=0.0.0.0:8080"
CMD ["go", "run", "main.go"]
