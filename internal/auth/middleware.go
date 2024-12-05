package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/SamSyntax/Gator/internal/database"
	"github.com/SamSyntax/Gator/internal/utils"
)

type handlerFunc func(s *utils.State, cmd utils.Command, user database.User) error

func MiddlewareLoggedIn(handler handlerFunc) func(*utils.State, utils.Command) error {
	return func(s *utils.State, cmd utils.Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Cfg.CURRENT_USER_NAME)
		if err != nil {
			return errors.Join(fmt.Errorf("Couldn't find user in database"), err)
		}
		err = handler(s, cmd, user)
		if err != nil {
			return errors.Join(fmt.Errorf("Failed to create middleware handler: "), err)
		}
		return nil
	}
}
