package search

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thss-cercis/cercis-server/api"
	"github.com/thss-cercis/cercis-server/db"
	"github.com/thss-cercis/cercis-server/db/user"
)

func SearchUser(c *fiber.Ctx) error {
	req := new(struct {
		ID       int64  `json:"id" form:"id" validate:"required_without_all=Mobile NickName"`
		Mobile   string `json:"mobile" form:"mobile" validate:"omitempty,phone_number"`
		NickName string `json:"nickname" form:"nickname"`
		Offset   int64  `json:"offset" form:"offset"`
		Limit    int64  `json:"limit" form:"limit"`
	})

	if ok, err := api.ParamParserWrap(c, req); !ok {
		return err
	}

	if ok, err := api.ValidateWrap(c, req); !ok {
		return err
	}

	type resType struct {
		ID       int64  `json:"id"`
		Mobile   string `json:"mobile,omitempty"`
		NickName string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	userToResType := func(u *user.User) resType {
		if u.AllowShowPhone {
			return resType{
				ID:       u.ID,
				Mobile:   u.Mobile,
				NickName: u.NickName,
				Avatar:   u.Avatar,
			}
		} else {
			return resType{
				ID:       u.ID,
				NickName: u.NickName,
				Avatar:   u.Avatar,
			}
		}
	}

	users := make([]resType, 0)
	//reqID, _ := strconv.Atoi(req.ID)
	if req.ID != 0 {
		u, err := user.GetUserByID(db.GetDB(), req.ID)
		if err == nil && u != nil {
			users = append(users, userToResType(u))
		}
	} else if req.Mobile != "" {
		u, err := user.GetUserByMobile(db.GetDB(), req.Mobile)
		if err == nil && u != nil && u.AllowSearchByPhone {
			users = append(users, userToResType(u))
		}
	} else if req.NickName != "" {
		us, err := user.GetUserLikeNickNameWithOffsetAndLimit(db.GetDB(), req.NickName, req.Offset, req.Limit)
		if err == nil {
			for _, u := range us {
				if u.AllowSearchByName {
					users = append(users, userToResType(&u))
				}
			}
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(api.BaseRes{Code: api.CodeFailure, Msg: "参数异常"})
	}

	return c.Status(fiber.StatusOK).JSON(api.BaseRes{Code: api.CodeSuccess, Msg: api.MsgSuccess, Payload: struct {
		Users []resType `json:"users"`
	}{
		Users: users,
	}})
}
