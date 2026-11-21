package authverify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitee.com/matkurban/openimsdk-tools/errs"
	"gitee.com/matkurban/openimsdk-tools/mcontext"
	"gitee.com/matkurban/openimsdk-tools/utils/datautil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/openimsdk/protocol/constant"
	"server/v3/pkg/common/servererrs"
)

const (
	hoursOneDay  = 24
	secondBefore = 5
)

type Claims struct {
	UserID     string
	PlatformID int // login platform
	jwt.RegisteredClaims
}

func Secret(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrTokenMalformed.WrapMsg("unexpected signing method")
		}
		return []byte(secret), nil
	}
}

func BuildClaims(uid string, platformID int, ttl int64) Claims {
	now := time.Now()
	before := now.Add(-time.Second * time.Duration(secondBefore))
	return Claims{
		UserID:     uid,
		PlatformID: platformID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl*hoursOneDay) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(before),
		},
	}
}

func CreateToken(uid string, platformID int, ttl int64, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, BuildClaims(uid, platformID, ttl))
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errs.WrapMsg(err, "token.SignedString")
	}
	return tokenString, nil
}

func GetClaimFromToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		Secret(secret),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err == nil {
		if claims, ok := token.Claims.(*Claims); ok && claims != nil && token.Valid {
			return claims, nil
		}
		return nil, errs.WrapMsg(errs.ErrTokenUnknown, "claims unknown", "token", tokenString)
	}
	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, errs.WrapMsg(errs.ErrTokenMalformed, "jwt parse error", "token", tokenString)
	case errors.Is(err, jwt.ErrTokenExpired):
		return nil, errs.WrapMsg(errs.ErrTokenExpired, "jwt parse error", "token", tokenString)
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, errs.WrapMsg(errs.ErrTokenNotValidYet, "jwt parse error", "token", tokenString)
	default:
		return nil, errs.WrapMsg(errs.NewCodeError(errs.TokenUnknownError, err.Error()), "jwt parse error", "token", tokenString)
	}
}

func CheckAdmin(ctx context.Context) error {
	if IsAdmin(ctx) {
		return nil
	}
	return servererrs.ErrNoPermission.WrapMsg(fmt.Sprintf("user %s is not admin userID", mcontext.GetOpUserID(ctx)))
}

//func IsManagerUserID(opUserID string, imAdminUserID []string) bool {
//	return datautil.Contain(opUserID, imAdminUserID...)
//}

func CheckUserIsAdmin(ctx context.Context, userID string) bool {
	return datautil.Contain(userID, GetIMAdminUserIDs(ctx)...)
}

func CheckSystemAccount(level int32) bool {
	return level >= constant.AppAdmin
}

const (
	CtxAdminUserIDsKey = "CtxAdminUserIDsKey"
)

func WithIMAdminUserIDs(ctx context.Context, imAdminUserID []string) context.Context {
	return context.WithValue(ctx, CtxAdminUserIDsKey, imAdminUserID)
}

func GetIMAdminUserIDs(ctx context.Context) []string {
	imAdminUserID, _ := ctx.Value(CtxAdminUserIDsKey).([]string)
	return imAdminUserID
}

func IsAdmin(ctx context.Context) bool {
	return IsTempAdmin(ctx) || IsSystemAdmin(ctx)
}

func CheckAccess(ctx context.Context, ownerUserID string) error {
	if mcontext.GetOpUserID(ctx) == ownerUserID {
		return nil
	}
	if IsAdmin(ctx) {
		return nil
	}
	return servererrs.ErrNoPermission.WrapMsg("ownerUserID", ownerUserID)
}

func CheckAccessIn(ctx context.Context, ownerUserIDs ...string) error {
	opUserID := mcontext.GetOpUserID(ctx)
	for _, userID := range ownerUserIDs {
		if opUserID == userID {
			return nil
		}
	}
	if IsAdmin(ctx) {
		return nil
	}
	return servererrs.ErrNoPermission.WrapMsg("opUser in ownerUserIDs")
}

var tempAdminValue = []string{"1"}

const ctxTempAdminKey = "ctxImTempAdminKey"

func WithTempAdmin(ctx context.Context) context.Context {
	keys, _ := ctx.Value(constant.RpcCustomHeader).([]string)
	if datautil.Contain(ctxTempAdminKey, keys...) {
		return ctx
	}
	if len(keys) > 0 {
		temp := make([]string, 0, len(keys)+1)
		temp = append(temp, keys...)
		keys = append(temp, ctxTempAdminKey)
	} else {
		keys = []string{ctxTempAdminKey}
	}
	ctx = context.WithValue(ctx, constant.RpcCustomHeader, keys)
	return context.WithValue(ctx, ctxTempAdminKey, tempAdminValue)
}

func IsTempAdmin(ctx context.Context) bool {
	values, _ := ctx.Value(ctxTempAdminKey).([]string)
	return datautil.Equal(tempAdminValue, values)
}

func IsSystemAdmin(ctx context.Context) bool {
	return datautil.Contain(mcontext.GetOpUserID(ctx), GetIMAdminUserIDs(ctx)...)
}
