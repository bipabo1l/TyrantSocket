package utils

// 用于返回标准接口数据
func AjaxReturn(data interface{}, info string, status int) (map[string]interface{}) {
	return map[string]interface{}{
		"code":  0,
		"count": status,
		"data":  data,
		"msg":   "",
	}
}
