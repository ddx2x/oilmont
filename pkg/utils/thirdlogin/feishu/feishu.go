package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ddx2x/oilmont/pkg/resource/iam"
	third "github.com/ddx2x/oilmont/pkg/utils/thirdlogin"
)

const (
	appId         = "appId"
	appSecret     = "Secret"
	tenantURL     = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	accountURL    = "https://open.feishu.cn/open-apis/authen/v1/access_token"
	DepartmentURL = "https://open.feishu.cn/open-apis/contact/v3/departments/%s/children"
	MemberURL     = "https://open.feishu.cn/open-apis/contact/v3/users?department_id=%s"
)

type FeiShu struct {
	TenantAccessToken string `json:"tenant_access_token"`
	ExpireTime        time.Time
}

func (t *FeiShu) GetAccountAccessToken(code string) (*third.Data, error) {
	if t.ExpireTime.Before(time.Now()) {
		if err := t.GetTenantAccessToken(); err != nil {
			return nil, err
		}
	}

	reqBody := map[string]interface{}{
		"grant_type": "authorization_code",
		"code":       code,
	}
	reqData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", accountURL, bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.TenantAccessToken))

	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	account := &third.AccountResponse{}

	err = json.Unmarshal(respData, account)
	if err != nil {
		return nil, err
	}

	if account.Code != 0 {
		return nil, fmt.Errorf("get account data error")
	}

	return &account.Data, nil
}

func (t *FeiShu) RefreshLIZIUser() error {
	if t.ExpireTime.Before(time.Now()) {
		if err := t.GetTenantAccessToken(); err != nil {
			return err
		}
	}

	businessGroups := make([]iam.BusinessGroup, 0)
	users := make([]iam.User, 0)

	err := t.loopDepartment("0", "", &businessGroups, &users)
	if err != nil {
		return err
	}
	return nil
}

func (t *FeiShu) loopDepartment(id string, pageToken string, businessGroups *[]iam.BusinessGroup, users *[]iam.User) error {
	requestURL := fmt.Sprintf(DepartmentURL, id)
	if pageToken != "" {
		requestURL = fmt.Sprintf("%s?page_token=%s", requestURL, pageToken)
	}

	req, err := http.NewRequest("GET", requestURL, bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.TenantAccessToken))

	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	departments := &third.DepartmentResponse{}

	err = json.Unmarshal(respData, departments)
	if err != nil {
		return err
	}

	if departments.Data.HasMore {
		err = t.loopDepartment(id, departments.Data.PageToken, businessGroups, users)
		if err != nil {
			return err
		}
	}

	for _, item := range departments.Data.Items {

		if err = t.getDepartmentMember(item.OpenDepartmentID, users); err != nil {
			return err
		}

		if err = t.loopDepartment(item.OpenDepartmentID, "", businessGroups, users); err != nil {
			return err
		}

	}

	return nil
}

func (t *FeiShu) getDepartmentMember(id string, users *[]iam.User) error {
	requestURL := fmt.Sprintf(MemberURL, id)

	req, err := http.NewRequest("GET", requestURL, bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.TenantAccessToken))

	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	departments := &third.DepartmentResponse{}

	err = json.Unmarshal(respData, departments)
	if err != nil {
		return err
	}

	return nil
}

func (t *FeiShu) GetTenantAccessToken() error {
	reqBody := map[string]interface{}{
		"app_id":     appId,
		"app_secret": appSecret,
	}
	reqData, _ := json.Marshal(reqBody)

	resp, err := http.Post(tenantURL, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respData, t)
	if err != nil {
		return err
	}
	t.ExpireTime = time.Now().Add(time.Minute * 100)

	return nil
}

func NewFeiShu() third.IThirdPartLogin {
	return &FeiShu{}
}
