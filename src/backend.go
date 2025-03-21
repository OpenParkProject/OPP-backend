//go:generate oapi-codegen -generate types,gin-server -o api/api.gen.go -package api api/openapi.yaml

package main

import (
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    api "OPP/backend/api"
)

type TestServer struct {
    tests map[int64]api.Test
    nextID int64
    mutex sync.Mutex
}

func NewTestServer() *TestServer {
    return &TestServer{
        tests: make(map[int64]api.Test),
        nextID: 1,
    }
}

// AddTest implements the POST /test endpoint
func (s *TestServer) AddTest(c *gin.Context) {
    // Parse request body
    var newTest api.Test
    if err := c.ShouldBindJSON(&newTest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if newTest.Id == nil {
        newTest.Id = &s.nextID
        s.nextID++
    }

    s.tests[*newTest.Id] = newTest

    c.JSON(http.StatusCreated, newTest)
}

func main() {
    r := gin.Default()
    
    server := NewTestServer()
    
    // Register API handlers
    api.RegisterHandlersWithOptions(r, server, api.GinServerOptions{
        BaseURL: "",  // No prefix for routes
        ErrorHandler: func(c *gin.Context, err error, statusCode int) {
            c.JSON(statusCode, gin.H{"error": err.Error()})
        },
    })
    
    fmt.Println("OPP Backend starting on :10020")
    log.Fatal(r.Run(":10020"))
}