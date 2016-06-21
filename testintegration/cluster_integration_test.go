package testintegration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/smancke/guble/client"
	"github.com/smancke/guble/gubled"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server"
	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"testing"
	"time"
	//"github.com/smancke/guble/testutil"
)

func createService(storagePath, nodeID, nodePort, listenPort string, remotes string) *server.Service {
	os.Args = []string{os.Args[0],
		"--log", "debug",
		"--listen", listenPort,
		"--storage-path", storagePath,
		"--node-id", nodeID,
		"--health", "",
		"--node-port", nodePort,
		remotes,
	}

	kingpin.Parse()
	service := gubled.StartService()
	return service
}

func Test_Cluster(t *testing.T) {
	a := assert.New(t)
	//defer testutil.EnableDebugForMethod()()

	service1 := createService("/tmp/s1", "1", "10000", "127.0.0.1:8080", "tcp://127.0.0.1:10000")
	a.NotNil(service1)

	service2 := createService("/tmp/s2", "2", "10001", "127.0.0.1:8081", "tcp://127.0.0.1:10000")
	a.NotNil(service2)

	client1, err1 := client.Open("ws://127.0.0.1:8081/stream/user/user1", "http://localhost", 1, false)
	assert.NoError(t, err1)

	client2, err2 := client.Open("ws://127.0.0.1:8080/stream/user/user2", "http://localhost", 1, false)
	assert.NoError(t, err2)

	err1 = client1.Subscribe("/foo")
	a.NoError(err1)

	err2 = client2.Subscribe("/testTopic")
	a.NoError(err2)

	err := client1.Send("/testTopic", "xyz", "{}")
	a.NoError(err)

	timeoutValue := time.Second

	//see if the message arrived at the other client
	select {
	case incomingMessage := <-client2.Messages():
		logger.WithFields(log.Fields{
			"nodeID":            incomingMessage.NodeID,
			"path":              incomingMessage.Path,
			"incomingMsgUserId": incomingMessage.UserID,
			"msg":               incomingMessage.BodyAsString(),
		}).Info("Client2 Received:")

		a.Equal(protocol.Path("/testTopic"), incomingMessage.Path)
		a.Equal("user1", incomingMessage.UserID)
		a.Equal("xyz", incomingMessage.BodyAsString())

	case <-time.After(timeoutValue):
		a.FailNow("No Message was received on second client until timeout")
	}

	// stop the cluster
	err1 = service1.Stop()
	err2 = service2.Stop()
	a.NoError(err1)
	a.NoError(err2)
}
