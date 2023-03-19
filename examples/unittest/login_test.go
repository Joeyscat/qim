package unittest

import (
	"testing"
	"time"

	"github.com/joeyscat/qim/examples/dialer"
	"github.com/stretchr/testify/assert"
)

const wsurl = "ws://localhost:8000"

func Test_login(t *testing.T) {
	cli, err := dialer.Login(wsurl, "u1")
	assert.NoError(t, err)
	time.Sleep(time.Second * 2)
	cli.Close()
}
