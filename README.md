This is a Go project that I am working on just to write something in Go. The system has an agent running in the cloud that can communicate back down to the server to search and download movies for a movie server. This was not designed to help facilitate the ability to illegally download movies and I do NOT condone any such behavior. I started this project to learn Go and to attempt to use all the different pieces of the language with interfaces, structs, readers/writers, as well as channels and goroutines.

*Still to do*
Implement a client ui library to easily make calls and see search results.

```
docker build -t pmd-server:dev \
    --build-arg PRIVATE_ID_KEY="`cat ~/.ssh/id_rsa`"\
    --build-arg VERSION="`git describe --abbrev=0 --tags | sed 's/v//g'`" \
    --build-arg BUILD="`git rev-parse --short HEAD`" \
    .

docker save pmd-server:dev | gzip > pmd-server.tar.gz
```