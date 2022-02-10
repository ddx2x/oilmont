package thirdlogin

type IThirdPartLogin interface {
	GetAccountAccessToken(code string) (*Data, error)
}

type Data struct {
	Name      string `json:"name"`
	EnName    string `json:"en_name"`
	UnionID   string `json:"union_id"`
	Email     string `json:"email"`
	TenantKey string `json:"tenant_key"`
}

type AccountResponse struct {
	Code int  `json:"code"`
	Data Data `json:"data"`
}

type Department struct {
	HasMore bool `json:"has_more"`
	Items   []struct {
		ChatId       string `json:"chat_id" bson:"chat_id"`
		DepartmentID string `json:"department_id"`
		I18NName     struct {
			EnUs string `json:"en_us"`
			JaJp string `json:"ja_jp"`
			ZhCn string `json:"zh_cn"`
		} `json:"i18n_name"`
		LeaderUserID       string `json:"leader_user_id"`
		MemberCount        int    `json:"member_count"`
		Name               string `json:"name"`
		OpenDepartmentID   string `json:"open_department_id"`
		Order              string `json:"order"`
		ParentDepartmentID string `json:"parent_department_id"`
		Status             struct {
			IsDeleted bool `json:"is_deleted"`
		} `json:"status"`
	} `json:"items"`
	PageToken string `json:"page_token"`
}

type DepartmentResponse struct {
	Code int        `json:"code"`
	Data Department `json:"data"`
}

type Member struct {
	HasMore bool `json:"has_more"`
	Items   []struct {
		Avatar struct {
			Avatar240    string `json:"avatar_240"`
			Avatar640    string `json:"avatar_640"`
			Avatar72     string `json:"avatar_72"`
			AvatarOrigin string `json:"avatar_origin"`
		} `json:"avatar"`
		DepartmentIds   []string `json:"department_ids"`
		Email           string   `json:"email"`
		EnName          string   `json:"en_name"`
		IsTenantManager bool     `json:"is_tenant_manager"`
		JobTitle        string   `json:"job_title"`
		JoinTime        int      `json:"join_time"`
		LeaderUserID    string   `json:"leader_user_id"`
		MobileVisible   bool     `json:"mobile_visible"`
		Name            string   `json:"name"`
		OpenID          string   `json:"open_id"`
		Orders          []struct {
			DepartmentID    string `json:"department_id"`
			DepartmentOrder int    `json:"department_order"`
			UserOrder       int    `json:"user_order"`
		} `json:"orders"`
		Status struct {
			IsActivated bool `json:"is_activated"`
			IsFrozen    bool `json:"is_frozen"`
			IsResigned  bool `json:"is_resigned"`
		} `json:"status"`
		UnionID     string `json:"union_id"`
		UserID      string `json:"user_id"`
		WorkStation string `json:"work_station"`
	} `json:"items"`
}

type MemberResponse struct {
	Code int    `json:"code"`
	Data Member `json:"data"`
}
