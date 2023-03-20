package dialer

import (
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/websocket"
	"github.com/joeyscat/qim/wire/token"
	"go.uber.org/zap"
)

func Login(wsurl, account string, appSecrets ...string) (qim.Client, error) {
	lg, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	cli := websocket.NewClient(account, "unittest", lg, websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
		secret = appSecrets[0]
	}
	cli.SetDialer(&ClientDialer{
		AppSecret: secret,
	})
	err = cli.Connect(wsurl)
	if err != nil {
		return nil, err
	}

	return cli, nil
}
