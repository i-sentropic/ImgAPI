# ImgAPI

Images are stored on a local instance of MongoDB. The server listens for HTTP requests over 8080 and gRPC requests over 8089, using a go routine to run both concurrently.

There are 4 packages:
1. db - Initialising the connection to the instance of MongoDB
2. proto - gRPC requests
3. routes - HTTP requests
4. src - image transformation, db operations and structs for JSON unmarshaling

## Testing
Testing is done through a series of gRPC client requests and http client requests located in the client directory. 
