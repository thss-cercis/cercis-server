package ws

// WriteToUser 将信息写给某个 user 的所有 session
func WriteToUser(userID int64, v interface{}) error {
	cons := GetConnByUserID(userID)
	var errRet error = nil
	for _, conn := range cons {
		if conn == nil {
			continue
		}
		if err := conn.Write(v); err != nil {
			errRet = err
		}
	}
	return errRet
}
