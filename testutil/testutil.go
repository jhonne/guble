package testutil

import (
	"encoding/json"
	_ "net/http/pprof"

	"github.com/Bogh/gcm"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution/health"
	"github.com/golang/mock/gomock"
	"github.com/smancke/guble/server/connector"
	"github.com/stretchr/testify/assert"

	"net/http"
	"os"
	"testing"
	"time"
)

// MockCtrl is a gomock.Controller to use globally
var MockCtrl *gomock.Controller

func init() {
	// disable error output while testing
	// because also negative tests are tested
	log.SetLevel(log.ErrorLevel)
}

// NewMockCtrl initializes the `MockCtrl` package var and returns a method to
// finish the controller when test is complete
// **Important**: Don't forget to call the returned method at the end of the test
// Usage:
// 		ctrl, finish := test_util.NewMockCtrl(t)
// 		defer finish()
func NewMockCtrl(t *testing.T) (*gomock.Controller, func()) {
	MockCtrl = gomock.NewController(t)
	return MockCtrl, func() { MockCtrl.Finish() }
}

// EnableDebugForMethod enables debug-level output through the current test
// Usage:
//		testutil.EnableDebugForMethod()()
func EnableDebugForMethod() func() {
	reset := log.GetLevel()
	log.SetLevel(log.DebugLevel)
	return func() { log.SetLevel(reset) }
}

// EnableInfoForMethod enables info-level output through the current test
// Usage:
//		testutil.EnableInfoForMethod()()
func EnableInfoForMethod() func() {
	reset := log.GetLevel()
	log.SetLevel(log.InfoLevel)
	return func() { log.SetLevel(reset) }
}

// ExpectDone waits to receive a value in the doneChannel for at least a second
// or fails the test.
func ExpectDone(a *assert.Assertions, doneChannel chan bool) {
	select {
	case <-doneChannel:
		return
	case <-time.After(time.Second):
		a.Fail("timeout in expectDone")
	}
}

// ExpectPanic expects a panic (and fails if this does not happen).
func ExpectPanic(t *testing.T) {
	if r := recover(); r == nil {
		assert.Fail(t, "Expecting a panic but unfortunately it did not happen")
	}
}

// ResetDefaultRegistryHealthCheck resets the existing registry containing health-checks
func ResetDefaultRegistryHealthCheck() {
	health.DefaultRegistry = health.NewRegistry()
}

const (
	SuccessFCMResponse = `{
	   "multicast_id":3,
	   "success":1,
	   "failure":0,
	   "canonical_ids":0,
	   "results":[
	      {
	         "message_id":"da",
	         "registration_id":"rId",
	         "error":""
	      }
	   ]
	}`

	ErrorFCMResponse = `{
	   "multicast_id":3,
	   "success":0,
	   "failure":1,
       "error":"InvalidRegistration",
	   "canonical_ids":5,
	   "results":[
	      {
	         "message_id":"err",
	         "registration_id":"fcmCanonicalID",
	         "error":"InvalidRegistration"
	      }
	   ]
	}`
)

type FCMSender func(request connector.Request) (interface{}, error)

func (fcms FCMSender) Send(request connector.Request) (interface{}, error) {
	return fcms(request)
}

func CreateFcmSender(body string, doneC chan bool, to time.Duration) (connector.Sender, error) {
	response := new(gcm.Response)

	err := json.Unmarshal([]byte(body), response)
	if err != nil {
		return nil, err
	}

	return FCMSender(func(request connector.Request) (interface{}, error) {
		defer func() {
			doneC <- true
		}()
		<-time.After(to)
		return response, nil
	}), nil
}

//SkipIfShort skips a test if the `-short` flag is given to `go test`
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
}

//SkipIfDisabled skips a test if the GO_TEST_DISABLED environment variable is set to any value (when `go test` runs)
func SkipIfDisabled(t *testing.T) {
	if os.Getenv("GO_TEST_DISABLED") != "" {
		t.Skip("skipping disabled test.")
	}
}

func PprofDebug() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}
