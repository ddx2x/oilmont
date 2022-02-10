package dingding

import third "github.com/ddx2x/oilmont/pkg/utils/thirdlogin"

var _ third.IThirdPartLogin = &dingDing{}

type dingDing struct{}

func (d dingDing) GetAccountAccessToken(code string) (*third.Data, error) {
	//TODO implement me
	panic("implement me")
}

func NewDingDing() third.IThirdPartLogin {
	return &dingDing{}
}
