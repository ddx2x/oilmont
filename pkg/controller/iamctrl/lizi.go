package iamctrl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/resource/iam"
	"github.com/ddx2x/oilmont/pkg/utils/thirdlogin"
	"github.com/ddx2x/oilmont/pkg/utils/thirdlogin/feishu"
)

func (r *RBACController) loopDepartment(token, id, pageToken string) error {
	requestURL := fmt.Sprintf(feishu.DepartmentURL, id)
	if pageToken != "" {
		requestURL = fmt.Sprintf("%s?page_token=%s", requestURL, pageToken)
	}

	req, err := http.NewRequest("GET", requestURL, bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	departments := &thirdlogin.DepartmentResponse{}

	err = json.Unmarshal(respData, departments)
	if err != nil {
		return err
	}

	for _, item := range departments.Data.Items {

		if err = r.getDepartmentMember(token, item.OpenDepartmentID); err != nil {
			return err
		}
		if err = r.loopDepartment(token, item.OpenDepartmentID, ""); err != nil {
			return err
		}
	}
	if departments.Data.HasMore {
		err = r.loopDepartment(token, id, departments.Data.PageToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RBACController) getDepartmentMember(token, id string) error {
	requestURL := fmt.Sprintf(feishu.MemberURL, id)

	req, err := http.NewRequest("GET", requestURL, bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	members := &thirdlogin.MemberResponse{}

	err = json.Unmarshal(respData, members)
	if err != nil {
		return err
	}

	tenantDatabase := "default"

	for _, item := range members.Data.Items {
		user := iam.User{
			Metadata: core.Metadata{
				Name:      strings.Split(item.Email, "@")[0],
				Tenant:    tenantDatabase,
				Workspace: common.DefaultWorkspace,
			},
			Spec: iam.UserSpec{
				Email:  item.Email,
				EnName: item.EnName,
				CnName: item.Name,
				OpenID: item.OpenID,
				Avatar: iam.Avatar{
					Avatar72:     item.Avatar.Avatar72,
					Avatar240:    item.Avatar.Avatar240,
					Avatar640:    item.Avatar.Avatar640,
					AvatarOrigin: item.Avatar.AvatarOrigin,
				},
			},
		}
		if user.GetName() == "" {
			user.Name = strings.ReplaceAll(item.EnName, " ", "")
		}

		oldUser := iam.User{}
		err := r.Get(tenantDatabase, common.USER, user.GetName(), &oldUser, true)
		if err == datasource.NotFound {
			if _, createErr := r.Create(tenantDatabase, common.USER, &user); createErr != nil {
				return createErr
			}
		} else {
			user.Spec.Account = oldUser.Spec.Account
			if _, _, applyErr := r.Apply(tenantDatabase, common.USER, user.GetName(), &user, false); applyErr != nil {
				return applyErr
			}
		}
	}
	return nil
}
