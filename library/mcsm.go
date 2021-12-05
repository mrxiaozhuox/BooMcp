package library

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Bulk_Register_MCSM_User
// 本函数将批量为账号注册 MCSM 账号
// 用户注册的 `oid` 为用户信息在 MongoDB 中的唯一 uuid
// 用户注册的 `pwds` 是系统预先随机生成完毕的密码列表（每一个MCSM站点生成不一样的密码）
func BulkRegisterMcsmUser(config GeneralConfig, oid string, pwds map[string]string) map[string]bool {

	// 对于用户 Oid 进行处理（鉴于ObjectID生成的长度限制）
	// https://docs.mongodb.com/manual/reference/method/ObjectId/
	// 所以说用户在 MCSM 的账号为其本身在 BooMCP 注册时 ObjectID 的前 12 位
	if len(oid) > 12 {
		// 根据 ObjectID 生成的方式来看，前12位已经完全可以唯一了，所以说把后面的省略掉也行
		oid = oid[0:12]
	}

	var status map[string]bool = make(map[string]bool)

	for _, value := range config.MCSMConnect {

		// 缺少路径 `/` 自动补充上去
		if value.Domain[len(value.Domain)-1:] != "/" {
			value.Domain += "/"
		}

		// log.Println("用户注册 MCSM 请求：" + value.Domain + "api/create_user/?apikey=" + value.MasterToken)

		log.Println(pwds[value.Name])
		log.Println(len(pwds[value.Name]))
		log.Println(oid)
		log.Println(len(oid))

		resp, err := http.PostForm(
			value.Domain+"api/create_user/?apikey="+value.MasterToken,
			url.Values{
				"username": {oid},
				"password": {
					pwds[value.Name],
				},
				"serverlist": {
					"",
				},
			},
		)
		// 将错误状态存入错误列表中作记录
		if err != nil {
			// 出错了
			status[value.Name] = false
			log.Println("MCSM账号注册失败：" + value.Domain + "api/create_user/?apikey=" + value.MasterToken)
			continue
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			status[value.Name] = false
			log.Println("MCSM账号注册失败：" + value.Domain + "api/create_user/?apikey=" + value.MasterToken)
			continue
		}

		bodyStr := string(body)

		if bodyStr == "{\"status\":200}" {
			status[value.Name] = true
		} else {
			log.Println("MCSM账号注册失败：" + value.Domain + "api/create_user/?apikey=" + value.MasterToken)
			status[value.Name] = false
		}

	}

	return status
}
