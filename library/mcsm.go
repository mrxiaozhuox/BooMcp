package library

import (
	"net/http"
	"net/url"
)

// Bulk_Register_MCSM_User
// 本函数将批量为账号注册 MCSM 账号
// 用户注册的 `oid` 为用户信息在 MongoDB 中的唯一 uuid
// 用户注册的 `pwds` 是系统预先随机生成完毕的密码列表（每一个MCSM站点生成不一样的密码）
func BulkRegisterMcsmUser(config GeneralConfig, oid string, pwds map[string]string) map[string]bool {

	var status map[string]bool

	for _, value := range config.MCSMConnect {
		_, err := http.PostForm(
			value.Domain+"/api/create_user/?apikey="+value.ApiKey,
			url.Values{
				"username": {oid},
				"password": {
					pwds[value.Name],
				},
			},
		)
		// 将错误状态存入错误列表中作记录
		status[value.Name] = (err == nil)
	}

	return status
}
